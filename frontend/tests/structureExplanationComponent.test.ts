import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import StructureExplanation from '../src/components/StructureExplanation.vue'
import { t } from '../src/locales'
import { rockSaltResult } from './fixtures'

describe('StructureExplanation', () => {
  it('用当前语言解释晶体类型、结构型与配位数', () => {
    const wrapper = mount(StructureExplanation, { props: { result: rockSaltResult } })
    const text = wrapper.text()

    expect(text).toContain(t('explain_title'))
    expect(text).toContain(t('family_ionic'))
    expect(text).toContain(t('prototype_rock_salt'))
    expect(text).toContain(String(rockSaltResult.coordination_number))
  })

  it('给出晶胞形状与三条棱、三个夹角', () => {
    const wrapper = mount(StructureExplanation, { props: { result: rockSaltResult } })
    const text = wrapper.text()

    expect(text).toContain(t('cell_shape_cubic'))
    expect(text).toContain('a = 5.66 Å')
    expect(text).toContain('γ = 90°')
  })

  it('判断依据来自后端 rule，不在前端重写结论', () => {
    const wrapper = mount(StructureExplanation, { props: { result: rockSaltResult } })

    expect(wrapper.text()).toContain(t(rockSaltResult.rule.code, rockSaltResult.rule.params))
  })

  it('没有注意事项时整块隐藏', () => {
    const wrapper = mount(StructureExplanation, { props: { result: rockSaltResult } })

    expect(wrapper.text()).not.toContain(t('explain_caveats_title'))
  })

  it('有注意事项时逐条翻译', () => {
    const caveat = {
      code: 'caveat_near_boundary',
      params: { ratio: 0.414, boundary: 0.414 },
    }
    const wrapper = mount(StructureExplanation, {
      props: { result: { ...rockSaltResult, caveats: [caveat] } },
    })
    const text = wrapper.text()

    expect(text).toContain(t('explain_caveats_title'))
    expect(text).toContain(t(caveat.code, caveat.params))
  })

  it('后端以后新增晶体类型时，至少显示原始值而不是空白', () => {
    const wrapper = mount(StructureExplanation, {
      props: { result: { ...rockSaltResult, family: 'plasma' } },
    })

    expect(wrapper.text()).toContain('plasma')
  })
})
