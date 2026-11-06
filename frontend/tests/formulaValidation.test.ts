import { describe, expect, it } from 'vitest'

import { MAX_FORMULA_LENGTH, validateFormulaSyntax } from '../src/formulaValidation'

describe('validateFormulaSyntax', () => {
  it.each(['NaCl', 'CaF2', 'Ca(OH)2', 'C', 'Na2O', 'Ca3(PO4)2', 'Al2(SO4)3'])(
    '接受写法正确的化学式 %s',
    (formula) => {
      expect(validateFormulaSyntax(formula)).toBeNull()
    },
  )

  it('先去掉首尾空格，再判断写法', () => {
    expect(validateFormulaSyntax('  NaCl  ')).toBeNull()
  })

  it('长度上限与后端列宽一致', () => {
    expect(MAX_FORMULA_LENGTH).toBe(64)
    expect(validateFormulaSyntax('C'.repeat(MAX_FORMULA_LENGTH))).toBeNull()
    expect(validateFormulaSyntax('C'.repeat(MAX_FORMULA_LENGTH + 1))).toEqual({
      code: 'task_formula_too_long',
      params: { max_length: MAX_FORMULA_LENGTH },
    })
  })

  it.each(['', '   ', '()'])('把空化学式 %j 报成 formula_empty', (formula) => {
    expect(validateFormulaSyntax(formula)).toEqual({ code: 'formula_empty' })
  })

  it('指出第几个字符看不懂（位置从 1 开始数）', () => {
    expect(validateFormulaSyntax('NaCl2!')).toEqual({
      code: 'formula_invalid_character',
      params: { character: '!', position: 6 },
    })
    expect(validateFormulaSyntax('2Na')).toEqual({
      code: 'formula_invalid_character',
      params: { character: '2', position: 1 },
    })
    expect(validateFormulaSyntax('Na Cl')).toEqual({
      code: 'formula_invalid_character',
      params: { character: ' ', position: 3 },
    })
  })

  it.each(['Ca(OH2', 'CaOH)2', 'Ca(OH))2'])('括号不配对时给出 formula_unbalanced_parentheses', (formula) => {
    expect(validateFormulaSyntax(formula)).toEqual({
      code: 'formula_unbalanced_parentheses',
      params: { formula },
    })
  })

  it('下标写成 0 时报 formula_zero_count', () => {
    expect(validateFormulaSyntax('CaF0')).toEqual({
      code: 'formula_zero_count',
      params: { formula: 'CaF0' },
    })
  })
})
