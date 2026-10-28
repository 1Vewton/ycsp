// 界面文案的取用入口：组件里只写 tag，由这里按当前语言取文本。
//
// 规则（见根 README「多语言」与同目录 README）：
// - 默认中文；用户切换过就记住选择。
// - 当前语言没有这条文案时回退英文；两边都没有就原样显示 tag 并在控制台告警，不把缺翻译藏起来。

import { readonly, ref } from 'vue'

import { enMessages } from './en'
import { zhMessages } from './zh'

export type { MessageKey } from './en'

/** 界面语言。 */
export type Locale = 'zh' | 'en'

/** 默认语言：产品默认中文。 */
export const DEFAULT_LOCALE: Locale = 'zh'

/** 兜底语言：任何语言缺这条文案时都用英文。 */
export const FALLBACK_LOCALE: Locale = 'en'

/** 可切换的语言，顺序即界面上按钮的顺序。 */
export const AVAILABLE_LOCALES: readonly Locale[] = ['zh', 'en']

/** 语言自己的称呼，不随界面语言变化，所以不放进语言包。 */
export const LOCALE_NAMES: Record<Locale, string> = {
  zh: '中文',
  en: 'English',
}

/** 文案里 {name} 占位符要填的值。 */
export type TextParams = Record<string, unknown>

// 匹配 {name} 形式的占位符。
const PLACEHOLDER_PATTERN = /\{(\w+)\}/g

// 记住用户的选择；读不到（如隐私模式）就只在本次访问有效。
const STORAGE_KEY = 'ycsp_locale'

// 值可能缺失（缺翻译），所以类型里带上 undefined，查找时才能判断“没找到”。
const MESSAGE_TABLES: Record<Locale, Record<string, string | undefined>> = {
  zh: zhMessages,
  en: enMessages,
}

const activeLocale = ref<Locale>(readStoredLocale())
applyDocumentLanguage(activeLocale.value)

/** 当前语言（只读）。在模板或计算属性里读取它会随切换自动更新。 */
export const currentLocale = readonly(activeLocale)

/** 切换界面语言，并记住选择（下次打开仍是这个语言）。 */
export function setLocale(locale: Locale): void {
  activeLocale.value = locale
  applyDocumentLanguage(locale)
  storeLocale(locale)
}

/**
 * 取 tag 对应的文案，并把 {name} 占位符换成 params 里的值。
 *
 * 查找顺序：当前语言 → 英文 → 原样返回 tag（同时 console.warn，避免缺翻译被悄悄藏起来）。
 */
export function t(tag: string, params: TextParams = {}): string {
  const template = findTemplate(tag)
  if (template === undefined) {
    console.warn(`[ycsp] 缺少文案 tag：${tag}`)
    return tag
  }
  return fillPlaceholders(template, params)
}

function findTemplate(tag: string): string | undefined {
  for (const locale of [activeLocale.value, FALLBACK_LOCALE]) {
    const template = MESSAGE_TABLES[locale][tag]
    if (template !== undefined) {
      return template
    }
  }
  return undefined
}

function fillPlaceholders(template: string, params: TextParams): string {
  return template.replace(PLACEHOLDER_PATTERN, (placeholder: string, name: string): string => {
    return name in params ? String(params[name]) : placeholder
  })
}

function readStoredLocale(): Locale {
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    return isLocale(stored) ? stored : DEFAULT_LOCALE
  } catch {
    // 读不到 localStorage 也不影响使用，只是记不住选择。
    return DEFAULT_LOCALE
  }
}

function storeLocale(locale: Locale): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, locale)
  } catch {
    // 存不进去就算了：本次访问仍然是切换后的语言。
  }
}

function isLocale(value: string | null): value is Locale {
  return value === 'zh' || value === 'en'
}

// 同步 <html lang>：屏幕阅读器与浏览器的翻译功能都靠它判断页面语言。
function applyDocumentLanguage(locale: Locale): void {
  document.documentElement.lang = locale === 'zh' ? 'zh-CN' : 'en'
}
