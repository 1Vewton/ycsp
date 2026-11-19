> [English](./README.md) | 简体中文

# 演示文稿（slides）

一份面向青少年的科普型 Typst 幻灯片：介绍 ycsp 是什么、学生怎么用，以及“从化学式到晶胞”的教学算法。它是演示材料，不参与平台运行。

## 文件

| 文件 | 作用 |
| --- | --- |
| `ycsp-slides.typ` | Typst 源文件：16:9，22 页，不依赖外部包 |
| `ycsp-slides.pdf` | 已编译好的演示文稿，可直接放映 |

## 编译

需要 Typst 0.15+ 和中文字体。源文件优先使用 `Microsoft YaHei`，找不到时回退 `SimHei`；Windows 自带这两款字体。

```powershell
typst compile ycsp-slides.typ ycsp-slides.pdf
```

如需导出 PNG（例如导入 PowerPoint）：

```powershell
typst compile ycsp-slides.typ "slide-{n}.png" --format png --ppi 110
```

Typst 输出的是 PDF，不是 `.pptx`。确实需要可编辑的 PowerPoint 文件时，把 PDF 或 PNG 导入 PowerPoint 即可。

## 内容结构

| 页码 | 内容 |
| --- | --- |
| 1–4 | 用日常语言解释晶体与 CSP；项目定位；学生使用流程；三层架构 |
| 5–14 | 算法：五步流水线、化学式解析、晶体家族判断、硬球半径比、NaCl 完整算例、规则 vs 实验、萤石/反萤石、金属、共价晶体 |
| 15–19 | 结果解读（`rule` / `caveats`）、工程保障、支持范围、教学价值 |
| 20–22 | 结尾与可尝试的例子 |

## 修改

- 文案为中文，技术术语首次出现时都给出通俗解释。
- 颜色、版式组件与页面模板都在 `ycsp-slides.typ` 顶部。
- 页数 = 封面 + `#slide` 与 `#section-page` 的调用数量，没有外部主题。

项目整体说明见根 [README](../README.zh-CN.md)，算法细节见 [compute/README.zh-CN.md](../compute/README.zh-CN.md)。
