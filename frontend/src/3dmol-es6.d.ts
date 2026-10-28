// 3Dmol 的 npm 入口是 UMD 构建（main: build/3Dmol.js），在打包器里拿命名导出不稳定。
// 这里显式使用它附带的 ES 模块构建，并把类型转接到包自带的声明。
declare module '3dmol/build/3Dmol.es6-min.js' {
  export * from '3dmol'
}
