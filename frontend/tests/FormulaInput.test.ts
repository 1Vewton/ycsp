import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import FormulaInput from '../src/components/FormulaInput.vue'
import type { FormulaProblem } from '../src/formulaValidation'
import { t } from '../src/locales'

function mountInput(modelValue = '', problem: FormulaProblem | null = null) {
  return mount(FormulaInput, { props: { modelValue, problem } })
}

describe('FormulaInput', () => {
  it('还没输入时给出“怎么写”的提示', () => {
    const wrapper = mountInput()

    expect(wrapper.find('.formula-feedback').text()).toBe(t('formula_input_hint'))
    expect(wrapper.find('.formula-field').attributes('aria-invalid')).toBe('false')
    expect(wrapper.find('.formula-field').classes()).not.toContain('formula-field-invalid')
  })

  it('写法正确时告诉学生可以继续', () => {
    const wrapper = mountInput('CaF2')

    expect(wrapper.find('.formula-feedback').text()).toBe(t('formula_syntax_ok'))
    expect(wrapper.find('.formula-feedback').classes()).toContain('formula-feedback-success')
  })

  it('页面传来的问题按当前语言显示，并标红输入框', () => {
    const problem: FormulaProblem = {
      code: 'formula_invalid_character',
      params: { character: '!', position: 6 },
    }
    const wrapper = mountInput('CaF2!', problem)

    expect(wrapper.find('.formula-feedback').text()).toBe(t(problem.code, problem.params))
    expect(wrapper.find('.formula-field').attributes('aria-invalid')).toBe('true')
    expect(wrapper.find('.formula-field').classes()).toContain('formula-field-invalid')
  })

  it('输入时把新值交给页面（v-model）', async () => {
    const wrapper = mountInput()

    await wrapper.find('.formula-field').setValue('NaCl')

    expect(wrapper.emitted('update:modelValue')).toEqual([['NaCl']])
  })

  it('label / aria-describedby 与输入框、提示文字互相对应', () => {
    const wrapper = mountInput()
    const field = wrapper.find('.formula-field')

    expect(wrapper.find('label').attributes('for')).toBe(field.attributes('id'))
    expect(field.attributes('aria-describedby')).toBe(
      wrapper.find('.formula-feedback').attributes('id'),
    )
  })
})
