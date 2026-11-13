import { beforeEach, describe, expect, it, vi } from 'vitest'

import { enMessages } from '../src/locales/en'
import { zhMessages } from '../src/locales/zh'

// locales/index.ts 在模块加载时读 localStorage 并写 <html lang>，
// 所以每个用例都重置模块，相当于“重新打开一次页面”。
async function loadLocales() {
  vi.resetModules()
  return import('../src/locales')
}

beforeEach(() => {
  window.localStorage.clear()
  document.documentElement.lang = ''
})

describe('默认语言与记忆', () => {
  it('没有记录时默认中文，并同步 <html lang>', async () => {
    const { currentLocale, t } = await loadLocales()

    expect(currentLocale.value).toBe('zh')
    expect(t('nav_home')).toBe(zhMessages.nav_home)
    expect(document.documentElement.lang).toBe('zh-CN')
  })

  it('读取上次选择并同步 <html lang>', async () => {
    window.localStorage.setItem('ycsp_locale', 'en')
    const { currentLocale, t } = await loadLocales()

    expect(currentLocale.value).toBe('en')
    expect(t('nav_home')).toBe(enMessages.nav_home)
    expect(document.documentElement.lang).toBe('en')
  })

  it('切换后写回 localStorage，重新加载仍是同一语言', async () => {
    const first = await loadLocales()
    first.setLocale('en')

    expect(window.localStorage.getItem('ycsp_locale')).toBe('en')
    expect(document.documentElement.lang).toBe('en')

    const second = await loadLocales()
    expect(second.currentLocale.value).toBe('en')
  })

  it('记录不合法时回退默认中文', async () => {
    window.localStorage.setItem('ycsp_locale', 'fr')
    const { currentLocale } = await loadLocales()

    expect(currentLocale.value).toBe('zh')
  })
})

describe('取词与占位符', () => {
  it('把 {name} 换成 params 里的值', async () => {
    const { t } = await loadLocales()
    const text = t('task_not_found', { task_id: 7 })

    expect(text).toContain('7')
    expect(text).not.toContain('{task_id}')
  })

  it('缺参数时保留占位符，开发时能发现漏传', async () => {
    const { t } = await loadLocales()

    expect(t('task_not_found')).toContain('{task_id}')
  })

  it('未收录的 tag 原样返回并告警，不悄悄显示空白', async () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
    const { t } = await loadLocales()

    expect(t('tag_that_does_not_exist')).toBe('tag_that_does_not_exist')
    expect(warn).toHaveBeenCalledOnce()
  })

  it('未收录的后端 code 显示 unknown_code，而不是内部编号', async () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
    const { t, tCode } = await loadLocales()

    const text = tCode('task_future_code')

    expect(text).toBe(t('unknown_code', { code: 'task_future_code' }))
    expect(text).toContain('task_future_code')
    expect(warn).toHaveBeenCalledOnce()
  })

  it('新增的晶胞形状、结构型与规则文案都已收录', async () => {
    const { t } = await loadLocales()
    const keys = [
      'cell_shape_tetragonal',
      'prototype_wurtzite',
      'prototype_rutile',
      'rule_wurtzite_experimental',
      'rule_rutile_experimental',
      'caveat_polymorph',
    ]

    for (const key of keys) {
      expect(t(key)).not.toBe(key)
    }
  })

  it('登录页会逐条显示的密码规则都有文案', async () => {
    const { t, tCode } = await loadLocales()
    const params = { min_length: 8, max_bytes: 72 }
    // 与 backend/internal/auth/policy.go 的 PasswordRule 取值一致（见 internal/auth/README.md）。
    const rules = [
      'password_rule_too_short',
      'password_rule_too_long',
      'password_rule_missing_letter',
      'password_rule_missing_digit',
      'password_rule_missing_symbol',
    ]

    for (const rule of rules) {
      expect(tCode(rule, params)).not.toBe(t('unknown_code', { code: rule }))
    }
  })
})
