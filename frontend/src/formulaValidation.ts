// 化学式“写法”校验：字符、括号、下标。
//
// 只做前端能独立判断的部分：元素是否收录、电荷是否守恒、这类晶体是否支持，
// 都由数据处理层判断（元素与电荷数据只在那边，见 compute/app/domain/composition.py）。
// 错误 code 与数据处理层同名，所以界面文案直接复用 locales 里的同一套 tag。

// 与后端 prediction_tasks.formula 的列宽一致（见 backend/internal/api/tasks.go）。
export const MAX_FORMULA_LENGTH = 64

// 字符范围：元素符号用 A-Z / a-z，下标用 0-9。
const UPPERCASE_FIRST = 'A'
const UPPERCASE_LAST = 'Z'
const LOWERCASE_FIRST = 'a'
const LOWERCASE_LAST = 'z'
const DIGIT_FIRST = '0'
const DIGIT_LAST = '9'
const OPENING_BRACKET = '('
const CLOSING_BRACKET = ')'

/** 一条可翻译的问题：code 决定说哪句话，params 是要填进句子的值。 */
export interface FormulaProblem {
  code: string
  params?: Record<string, unknown>
}

interface SubscriptRead {
  nextPosition: number
  problem: FormulaProblem | null
}

/**
 * 检查化学式的写法；没有问题返回 null。
 *
 * 空输入也返回 formula_empty，由页面决定显示成“怎么写”的提示还是错误。
 */
export function validateFormulaSyntax(rawFormula: string): FormulaProblem | null {
  const formula = rawFormula.trim()
  if (formula === '') {
    return { code: 'formula_empty' }
  }
  if (formula.length > MAX_FORMULA_LENGTH) {
    return { code: 'task_formula_too_long', params: { max_length: MAX_FORMULA_LENGTH } }
  }

  let position = 0
  let openGroupCount = 0
  let sawElement = false

  while (position < formula.length) {
    const character = formula[position]

    if (character === OPENING_BRACKET) {
      openGroupCount += 1
      position += 1
      continue
    }

    if (character === CLOSING_BRACKET) {
      if (openGroupCount === 0) {
        return unbalancedParentheses(formula)
      }
      openGroupCount -= 1
      const subscript = readSubscript(formula, position + 1)
      if (subscript.problem !== null) {
        return subscript.problem
      }
      position = subscript.nextPosition
      continue
    }

    const symbol = readElementSymbol(formula, position)
    if (symbol === null) {
      return {
        code: 'formula_invalid_character',
        params: { character, position: position + 1 },
      }
    }
    sawElement = true

    const subscript = readSubscript(formula, position + symbol.length)
    if (subscript.problem !== null) {
      return subscript.problem
    }
    position = subscript.nextPosition
  }

  if (openGroupCount > 0) {
    return unbalancedParentheses(formula)
  }
  // 只有空括号（如 "()"）时一个原子都没有，与数据处理层一样按空化学式处理。
  if (!sawElement) {
    return { code: 'formula_empty' }
  }
  return null
}

/** 读一个元素符号（一个大写字母，可跟一个小写字母）；读不到返回 null。 */
function readElementSymbol(formula: string, position: number): string | null {
  const first = formula[position]
  if (first < UPPERCASE_FIRST || first > UPPERCASE_LAST) {
    return null
  }
  const second = formula[position + 1]
  if (second !== undefined && second >= LOWERCASE_FIRST && second <= LOWERCASE_LAST) {
    return first + second
  }
  return first
}

/** 读一个下标（可以没有）；下标写成 0 时报错。 */
function readSubscript(formula: string, start: number): SubscriptRead {
  let end = start
  while (end < formula.length && isAsciiDigit(formula[end])) {
    end += 1
  }
  if (end === start) {
    return { nextPosition: start, problem: null }
  }

  const count = Number(formula.slice(start, end))
  if (count === 0) {
    return { nextPosition: end, problem: { code: 'formula_zero_count', params: { formula } } }
  }
  return { nextPosition: end, problem: null }
}

function isAsciiDigit(character: string): boolean {
  return character >= DIGIT_FIRST && character <= DIGIT_LAST
}

function unbalancedParentheses(formula: string): FormulaProblem {
  return { code: 'formula_unbalanced_parentheses', params: { formula } }
}
