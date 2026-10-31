// 检查后端 / 数据处理层的每个 code 是否都有中英文案。
//
// 语言包以 en.ts 为基准（zh.ts 由 `npm run type-check` 保证键一致）；
// 这里额外核对中英键，并让 CI 在“新增 code 但没补文案”时直接失败。
// 运行：在 frontend/ 下执行 `npm run check-codes`。

import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url))
const repositoryRoot = path.resolve(scriptDirectory, '..', '..')

const CODE_SOURCES = [
  { layer: 'backend', path: 'backend/internal/codes/codes.go' },
  { layer: 'compute', path: 'compute/app/constants/codes.py' },
]
const EN_MESSAGES = 'frontend/src/locales/en.ts'
const ZH_MESSAGES = 'frontend/src/locales/zh.ts'

// 解析失效时至少要有这么多条，避免“零差异”的假通过。
const MINIMUM_CODE_COUNT = 20
const MINIMUM_MESSAGE_COUNT = 100

const read = (relativePath) => readFileSync(path.join(repositoryRoot, relativePath), 'utf8')

const extractCodes = (source) =>
  [...source.matchAll(/=\s*"([a-z0-9_]+)"/g)].map((match) => match[1])

const extractMessageKeys = (source) =>
  new Set([...source.matchAll(/^  ([a-z0-9_]+):/gm)].map((match) => match[1]))

const enKeys = extractMessageKeys(read(EN_MESSAGES))
const zhKeys = extractMessageKeys(read(ZH_MESSAGES))

const codesByLayer = CODE_SOURCES.map((source) => ({
  layer: source.layer,
  codes: [...new Set(extractCodes(read(source.path)))],
}))
const allCodes = [...new Set(codesByLayer.flatMap((entry) => entry.codes))].sort()

const problems = []

if (enKeys.size < MINIMUM_MESSAGE_COUNT) {
  problems.push(`语言包只解析到 ${enKeys.size} 条，可能解析失败`)
}
if (allCodes.length < MINIMUM_CODE_COUNT) {
  problems.push(`只解析到 ${allCodes.length} 个 code，可能解析失败`)
}

for (const entry of codesByLayer) {
  if (entry.codes.length === 0) {
    problems.push(`${entry.layer}: 没有解析到 code`)
  }
  for (const code of entry.codes) {
    if (!enKeys.has(code)) {
      problems.push(`${entry.layer}: ${code} 缺少英文文案`)
    }
    if (!zhKeys.has(code)) {
      problems.push(`${entry.layer}: ${code} 缺少中文文案`)
    }
  }
}

for (const key of enKeys) {
  if (!zhKeys.has(key)) {
    problems.push(`语言包: zh.ts 缺少 ${key}`)
  }
}
for (const key of zhKeys) {
  if (!enKeys.has(key)) {
    problems.push(`语言包: en.ts 缺少 ${key}`)
  }
}

if (problems.length > 0) {
  console.error('错误提示文案检查未通过：')
  for (const problem of problems) {
    console.error(`  - ${problem}`)
  }
  console.error('请在 frontend/src/locales/en.ts 与 zh.ts 中补齐后重新运行。')
  process.exit(1)
}

console.log(`OK：${allCodes.length} 个后端 code 都有中英文案（语言包 ${enKeys.size} 条）。`)
