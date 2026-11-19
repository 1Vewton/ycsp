> English | [简体中文](./README.zh-CN.md)

# Presentation slides (slides)

A popular-science Typst deck about ycsp. It explains what crystal structure prediction (CSP) is, how a student uses the platform, and how the teaching algorithm turns a chemical formula into a unit cell. It is a presentation artifact, not part of the running application.

## Files

| File | Purpose |
| --- | --- |
| `ycsp-slides.typ` | Typst source: 16:9, 22 pages, no external packages |
| `ycsp-slides.pdf` | Compiled deck, ready to present |

## Build

Requires Typst 0.15+ and a Chinese font. The source tries `Microsoft YaHei` first and falls back to `SimHei`; both ship with Windows.

```powershell
typst compile ycsp-slides.typ ycsp-slides.pdf
```

To export PNG slides (for example, to import them into PowerPoint):

```powershell
typst compile ycsp-slides.typ "slide-{n}.png" --format png --ppi 110
```

Typst writes PDF, not `.pptx`. If an editable PowerPoint file is required, import the PDF or PNG output into PowerPoint.

## Outline

| Slides | Content |
| --- | --- |
| 1–4 | Crystal and CSP in everyday language; project positioning; student flow; three-layer architecture |
| 5–14 | Algorithm: five-step pipeline, formula parsing, family classification, hard-sphere radius ratio, NaCl worked example, rule vs experiment, fluorite/antifluorite, metals, covalent crystals |
| 15–19 | Reading results (`rule` / `caveats`), engineering guarantees, supported range, teaching value |
| 20–22 | Closing and examples to try |

## Editing

- All copy is Chinese and explains each technical term on first use.
- Colors, layout helpers, and the slide template live at the top of `ycsp-slides.typ`.
- Page count is the cover plus the number of `#slide` and `#section-page` calls; there is no external theme.

See the root [README](../README.md) for the project overview and [compute/README.md](../compute/README.md) for the algorithm details.
