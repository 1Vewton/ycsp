// 严格 CSP 不允许 eval / new Function；构建产物命中时，三维视图等代码可能被浏览器拦截。
// 这个脚本读取 dist 里的 JS，不参与浏览器构建。
import { readdir, readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

const assetsDirectory = fileURLToPath(new URL('../dist/assets/', import.meta.url))
const forbiddenPatterns = [
  { label: 'eval()', pattern: /\beval\s*\(/ },
  { label: 'new Function()', pattern: /\bnew\s+Function\b/ },
  { label: 'Function()', pattern: /\bFunction\s*\(/ },
]

let fileNames
try {
  fileNames = (await readdir(assetsDirectory)).filter((name) => name.endsWith('.js'))
} catch (error) {
  console.error(
    `check-csp: cannot read dist/assets; run "npm run build" first (${error.message}).`,
  )
  process.exit(1)
}

if (fileNames.length === 0) {
  console.error('check-csp: dist/assets contains no JS files; run "npm run build" first.')
  process.exit(1)
}

let failures = 0
for (const name of fileNames) {
  const source = await readFile(join(assetsDirectory, name), 'utf8')
  for (const { label, pattern } of forbiddenPatterns) {
    if (pattern.test(source)) {
      console.error(`check-csp: ${name} contains ${label}`)
      failures += 1
    }
  }
}

if (failures > 0) {
  console.error('check-csp: strict CSP would block the current build.')
  process.exit(1)
}

console.log(
  `OK: ${fileNames.length} built JS files contain no eval / Function constructor.`,
)
