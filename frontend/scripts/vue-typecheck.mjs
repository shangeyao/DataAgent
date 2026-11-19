// TypeScript 7 尚未提供稳定的编程 API，vue-tsc 通过官方兼容包 @typescript/typescript6 执行 SFC 类型检查。
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const shimRequire = createRequire(require.resolve('@typescript/typescript6/package.json'))
const tscPath = shimRequire.resolve('@typescript/typescript6/lib/tsc')
const { run } = require('vue-tsc')
run(tscPath)
