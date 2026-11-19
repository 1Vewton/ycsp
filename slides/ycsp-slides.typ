// ycsp 科普演示：从化学式到晶体结构。
// 只使用 Typst 原生能力，不依赖外部包，保证离线也能编译。

#set page(
  width: 33.867cm,
  height: 19.05cm,
  margin: 0pt,
  fill: rgb("#FBF8F1"),
)
#set text(
  font: ("Microsoft YaHei", "SimHei"),
  size: 18pt,
  fill: rgb("#14213D"),
  lang: "zh",
)
#set par(leading: 0.78em, justify: false)

#let ink = rgb("#14213D")
#let teal = rgb("#0F8B8D")
#let coral = rgb("#E76F51")
#let amber = rgb("#F4A261")
#let sky = rgb("#457B9D")
#let paper = rgb("#FBF8F1")
#let soft = rgb("#E8F3F1")
#let muted = rgb("#6B7280")
#let linegray = rgb("#D8D4C8")

// 一页标准内容页：标题、正文、页脚。
#let slide(title: none, kicker: none, body) = {
  pagebreak(weak: true)
  block(width: 100%, height: 100%, fill: paper, {
    place(top + left, rect(width: 100%, height: 0.18cm, fill: teal))
    block(width: 100%, height: 100%, inset: (x: 2.0cm, top: 1.0cm, bottom: 0.8cm), {
      grid(
        columns: 1,
        rows: (auto, 1fr, auto),
        row-gutter: 0.45cm,
        {
          if kicker != none [
            #text(size: 11pt, fill: teal, weight: "bold", tracking: 0.08em, kicker)
          ]
          if title != none [
            #block(above: 0.15cm, text(size: 28pt, weight: "bold", fill: ink, title))
          ]
        },
        align(top, body),
        {
          line(length: 100%, stroke: 0.5pt + linegray)
          grid(
            columns: (1fr, auto),
            align: horizon,
            text(size: 9pt, fill: muted, "ycsp · 青少年晶体结构预测教学平台"),
            context text(size: 9pt, fill: muted, counter(page).display("1 / 1", both: true)),
          )
        },
      )
    })
  })
}

// 轻量提示框：左侧有一条强调色。
#let note(body, accent: teal, fill: soft) = block(
  width: 100%,
  fill: fill,
  stroke: (left: 4pt + accent),
  radius: 4pt,
  inset: (x: 0.5cm, y: 0.32cm),
  body,
)

// 带标题的卡片。
#let card(title: none, body, accent: teal, fill: white, height: auto) = block(
  width: 100%,
  height: height,
  fill: fill,
  stroke: 0.8pt + accent.lighten(55%),
  radius: 7pt,
  inset: (x: 0.5cm, y: 0.4cm),
  {
    if title != none [
      #text(size: 15pt, weight: "bold", fill: accent, title)
      #v(0.12cm)
    ]
    body
  },
)

#let pill(body, fill: teal) = box(
  fill: fill,
  inset: (x: 7pt, y: 3pt),
  radius: 3pt,
  text(fill: white, size: 10.5pt, weight: "bold", body),
)

#let badge(number, accent: teal) = box(
  width: 0.72cm,
  height: 0.72cm,
  {
    place(circle(radius: 0.36cm, fill: accent))
    align(center + horizon, text(fill: white, size: 12pt, weight: "bold", number))
  },
)

// 编号步骤卡。
#let step(number, title, body, accent: teal) = block(
  width: 100%,
  fill: white,
  stroke: 0.8pt + accent.lighten(62%),
  radius: 6pt,
  inset: (x: 0.42cm, y: 0.24cm),
  {
    grid(
      columns: (auto, 1fr),
      column-gutter: 0.3cm,
      align: horizon,
      badge(number, accent: accent),
      text(size: 15pt, weight: "bold", fill: ink, title),
    )
    v(0.1cm)
    body
  },
)

// 流程小盒。
#let flow-box(title, subtitle, accent: teal) = block(
  width: 100%,
  height: 2.9cm,
  fill: white,
  stroke: 0.8pt + accent.lighten(58%),
  radius: 6pt,
  inset: (x: 0.22cm, y: 0.2cm),
  align(center + horizon, {
    text(size: 15pt, weight: "bold", fill: accent, title)
    linebreak()
    text(size: 10.5pt, fill: muted, subtitle)
  }),
)

#let arrow = text(size: 24pt, fill: teal)[→]

// NaCl 的二维切片：角上与中心是 Na，棱中是 Cl。
#let cell2d(size: 5.6cm) = box(width: size, height: size, {
  place(rect(width: 100%, height: 100%, stroke: 1.2pt + ink))
  let cells = range(9).map(index => {
    let row = calc.floor(index / 3)
    let col = index - row * 3
    align(center + horizon, circle(
      radius: 0.4cm,
      fill: if calc.rem(row + col, 2) == 0 { coral } else { sky },
    ))
  })
  grid(columns: (1fr,) * 3, rows: (1fr,) * 3, ..cells)
})

// 金刚石/闪锌矿的四面体键示意。
#let tetra(size: 5.6cm) = box(width: size, height: size * 0.9, {
  let center = (size * 0.5, size * 0.5)
  let points = (
    (size * 0.5, size * 0.06),
    (size * 0.08, size * 0.5),
    (size * 0.92, size * 0.5),
    (size * 0.5, size * 0.94),
  )
  for point in points {
    place(
      dx: center.at(0),
      dy: center.at(1),
      line(
        start: (0pt, 0pt),
        end: (point.at(0) - center.at(0), point.at(1) - center.at(1)),
        stroke: 3pt + sky,
      ),
    )
  }
  for point in points {
    place(dx: point.at(0) - 0.3cm, dy: point.at(1) - 0.3cm, circle(radius: 0.3cm, fill: sky))
  }
  place(dx: center.at(0) - 0.38cm, dy: center.at(1) - 0.38cm, circle(radius: 0.38cm, fill: coral))
})

// 金属堆积示意：fcc / bcc / hcp。
#let packing(kind, size: 4.0cm) = box(width: size, height: size, {
  let unit = size / 4
  let positions = if kind == "fcc" {
    ((1, 1), (2, 1), (3, 1), (1, 2), (2, 2), (3, 2), (1, 3), (2, 3), (3, 3))
  } else if kind == "bcc" {
    ((1, 1), (3, 1), (1, 3), (3, 3), (2, 2))
  } else {
    ((1.5, 1), (2.5, 1), (1, 2), (2, 2), (3, 2), (1.5, 3), (2.5, 3))
  }
  place(rect(width: 100%, height: 100%, stroke: 0.8pt + linegray, radius: 2pt))
  for point in positions {
    place(
      dx: point.at(0) * unit - 0.22cm,
      dy: point.at(1) * unit - 0.22cm,
      circle(radius: 0.22cm, fill: sky),
    )
  }
})

#let mini-card(title, body) = block(
  width: 100%,
  fill: white,
  stroke: 0.8pt + linegray,
  radius: 6pt,
  inset: (x: 0.25cm, y: 0.25cm),
  align(center, {
    body
    v(0.12cm)
    text(size: 10.5pt, weight: "bold", fill: ink, title)
  }),
)

// 分节页。
#let section-page(kicker, title, subtitle, accent: teal) = {
  pagebreak(weak: true)
  block(width: 100%, height: 100%, fill: accent, {
    place(top + right, dx: -2.2cm, dy: -2.2cm, circle(radius: 7.2cm, fill: white.transparentize(88%)))
    place(bottom + left, dx: -2.2cm, dy: 2.2cm, circle(radius: 5.2cm, fill: white.transparentize(92%)))
    block(width: 100%, height: 100%, inset: (x: 2.0cm, y: 1.6cm), {
      text(fill: white.transparentize(30%), size: 14pt, weight: "bold", kicker)
      v(0.35cm)
      text(fill: white, size: 38pt, weight: "bold", title)
      v(0.25cm)
      text(fill: white.transparentize(18%), size: 17pt, subtitle)
    })
  })
}

// 封面。
#block(width: 100%, height: 100%, fill: ink, {
  place(top + right, dx: -3.2cm, dy: -2.6cm, circle(radius: 6.2cm, fill: teal.transparentize(84%)))
  place(bottom + left, dx: -2.4cm, dy: 2.8cm, circle(radius: 4.6cm, fill: coral.transparentize(88%)))
  block(width: 100%, height: 100%, inset: (x: 2.0cm, y: 1.5cm), {
    grid(
      columns: 1,
      rows: (auto, 1fr, auto),
      {
        grid(
          columns: (auto, 1fr),
          align: horizon,
          box(fill: teal, inset: (x: 9pt, y: 4pt), radius: 3pt, text(fill: white, weight: "bold", size: 14pt)[ycsp]),
          text(fill: rgb("#B8C4D6"), size: 12pt, "晶体结构预测 · 青少年科普"),
        )
      },
      align(horizon, [
        #text(fill: white, size: 44pt, weight: "bold")[从一串化学式，]
        #linebreak()
        #text(fill: amber, size: 44pt, weight: "bold")[到一座晶体城市]
        #v(0.5cm)
        #text(fill: rgb("#D7DFEA"), size: 19pt)[用规则、硬球和实验数据，猜出物质内部的原子排列]
        #v(0.6cm)
        #grid(
          columns: (auto, auto, auto),
          gutter: 0.3cm,
          pill[化学式], pill(fill: coral)[晶体结构], pill(fill: sky)[三维解释],
        )
      ]),
      text(fill: rgb("#8FA0B8"), size: 10.5pt, "面向中学生的晶体结构预测（CSP）教学平台 · 2026"),
    )
  })
})


#slide(title: [晶体：原子排出的“城市”], kicker: [01 · 从生活出发])[
  #grid(
    columns: (1.08fr, 0.92fr),
    gutter: 0.8cm,
    [
      #text(size: 19pt, weight: "bold", fill: ink)[为什么盐是立方体，钻石那么硬？]
      #v(0.25cm)
      #text(size: 16pt)[食盐、冰糖、钻石、金属……它们的原子不是随便堆的，而是按同一套规则反复重复。]
      #v(0.25cm)
      #text(size: 16pt)[这种“重复排列”就是#text(weight: "bold", fill: teal)[晶体]；能代表整块晶体的最小重复单元，叫#text(weight: "bold", fill: teal)[晶胞]，可以想成一块瓷砖。]
      #v(0.35cm)
      #note[
        #text(weight: "bold", fill: teal)[CSP（晶体结构预测）]：给定化学成分，预测原子最可能怎么排。
        #v(0.1cm)
        #text(size: 14pt, fill: muted)[难在哪？同一成分可能有多种结构；结构由能量决定，而能量要算。]
      ]
    ],
    [
      #grid(
        columns: (1fr, 1fr),
        rows: (1fr, 1fr),
        gutter: 0.3cm,
        mini-card([食盐 NaCl], cell2d(size: 3.0cm)),
        mini-card([钻石 C], tetra(size: 3.2cm)),
        mini-card([金属 Cu], packing("fcc", size: 3.2cm)),
        mini-card([萤石 CaF#sub[2]], box(width: 3.2cm, height: 3.2cm, {
          let cells = range(9).map(index => {
            let row = calc.floor(index / 3)
            let col = index - row * 3
            let big = calc.rem(row, 2) == 0 and calc.rem(col, 2) == 0
            align(center + horizon, if big {
              circle(radius: 0.3cm, fill: coral)
            } else {
              circle(radius: 0.18cm, fill: sky)
            })
          })
          grid(columns: (1fr,) * 3, rows: (1fr,) * 3, ..cells)
        })),
      )
    ],
  )
]

#slide(title: [一个给青少年的 CSP 实验室], kicker: [02 · 项目定位])[
  #grid(
    columns: (1.05fr, 0.95fr),
    gutter: 0.8cm,
    [
      #grid(rows: (auto, auto, auto), row-gutter: 0.25cm,
        step("1", [输入一个化学式], [比如 NaCl、CaF#sub[2]、Cu。写法有问题时，输入框马上提示。]),
        step("2", [看到三维晶胞], [原子用小球表示，晶胞用线框表示；可以旋转、缩放、重置视角。], accent: coral),
        step("3", [读到解释与边界], [不只给结构名字，还说明“为什么”和“哪里不确定”。], accent: sky),
      )
      #v(0.2cm)
      #text(size: 13pt, fill: muted)[不需要编程，也不需要先学晶体学；术语第一次出现会用日常语言解释。]
      #v(0.2cm)
      #text(size: 12pt, fill: muted)[小词典：Å（埃）= 0.1 纳米，是原子尺度的长度单位。]
    ],
    [
      #card(title: [学生看到的界面])[
        #text(size: 11pt, fill: muted)[输入化学式]
        #v(0.08cm)
        #block(width: 100%, fill: paper, stroke: 0.8pt + linegray, radius: 4pt, inset: 0.2cm)[
          #text(size: 17pt)[CaF#sub[2]]
        ]
        #v(0.2cm)
        #box(fill: teal, inset: (x: 0.55cm, y: 0.16cm), radius: 4pt, text(fill: white, weight: "bold", size: 13pt)[开始预测])
        #v(0.3cm)
        #text(size: 11pt, fill: muted)[结果]
        #v(0.08cm)
        #text(size: 24pt, weight: "bold", fill: coral)[萤石型]
        #v(0.05cm)
        #text(size: 13pt)[配位数 8 · 晶胞棱长 5.38 Å]
        #v(0.3cm)
        #note(fill: white, [
          #text(size: 12pt)[“为什么是萤石型？因为 Ca#super[2+] 与 F#super[−] 的个数比是 1:2。”]
        ], accent: coral)
      ]
    ],
  )
]

#slide(title: [三步上手，不把学生挡在门外], kicker: [03 · 使用体验])[
  #grid(
    columns: (1fr, 1fr, 1fr),
    gutter: 0.5cm,
    step("1", [写化学式], [
      支持 NaCl、CaF#sub[2]、Ca(OH)#sub[2] 等写法；
      字符、括号、下标当场检查。
    ]),
    step("2", [点“开始预测”], [
      后台排队计算，页面显示“排队中 → 计算中 → 完成”；
      排队时还能取消，不会卡住页面。
    ], accent: coral),
    step("3", [看结构、读解释], [
      旋转三维晶胞；看到结构型、配位数、棱长，
      以及“为什么”和“哪里不确定”。
    ], accent: sky),
  )
  #v(0.5cm)
  #note[
    #text(weight: "bold")[登录不是前提。] 不登录也能预测；登录后才会把任务记到个人记录里，方便下次回看。
  ]
]

#slide(title: [三层架构：界面、数据、算法各司其职], kicker: [04 · 工程结构])[
  #grid(
    columns: (1.12fr, 0.88fr),
    gutter: 0.7cm,
    [
      #grid(rows: (auto, auto, auto, auto, auto), row-gutter: 0.12cm,
        card(title: [前端 · Vue + TypeScript], [
          学生界面、化学式写法校验、任务进度、三维结构展示。
        ]),
        align(center, text(size: 20pt, fill: teal)[↓]),
        card(title: [数据管理 · Go + Gin], [
          唯一对外接口、账号与会话、任务排队与落库、转发计算请求。
        ], accent: coral),
        align(center, text(size: 20pt, fill: teal)[↓]),
        card(title: [数据处理 · Python + FastAPI], [
          化学式解析、晶体类型判断、结构原型与晶胞计算。
        ], accent: sky),
      )
    ],
    [
      #card(title: [为什么这样分？], accent: coral)[
        #grid(rows: (auto, auto, auto), row-gutter: 0.25cm,
          text(size: 15pt)[#text(weight: "bold", fill: teal)[界面] 只关心“好不好用”。],
          text(size: 15pt)[#text(weight: "bold", fill: coral)[数据] 只关心“存得住、查得到”。],
          text(size: 15pt)[#text(weight: "bold", fill: sky)[算法] 只关心“算得对、说得清”。],
        )
        #v(0.25cm)
        #note[
          前端只和中间层说话；元素、电荷、错误码都只保留一份权威数据，避免三层各说各话。
        ]
      ]
    ],
  )
  #v(0.3cm)
  #note[请求流向：浏览器 → Go 数据管理 → Python 计算；结果再原路返回。]
]

#section-page([PART 01 · 算法篇], [算法怎么工作], [从化学式到晶胞的五步流水线])

#slide(title: [算法地图：五步流水线], kicker: [05 · 算法总览])[
  #grid(
    columns: (1fr, auto, 1fr, auto, 1fr, auto, 1fr, auto, 1fr),
    gutter: 0.16cm,
    align: horizon,
    flow-box([读化学式], [元素 → 个数]),
    arrow,
    flow-box([判家族], [金属 / 共价 / 离子], accent: coral),
    arrow,
    flow-box([选原型], [半径比 / 配比 / 实验表], accent: amber),
    arrow,
    flow-box([算晶胞], [硬球相切 / 实验参数], accent: sky),
    arrow,
    flow-box([出解释], [rule + caveats], accent: ink),
  )
  #v(0.45cm)
  #grid(
    columns: (1fr, 1fr, 1fr),
    gutter: 0.4cm,
    note([#text(weight: "bold")[先分类，再套规则。] 不同家族的规则完全不同。]),
    note([#text(weight: "bold")[不支持就明说。] 不硬给一个看似合理的答案。], accent: coral),
    note([#text(weight: "bold")[结果带解释。] 不只说“是什么”，还说“为什么”和“哪里不确定”。], accent: sky),
  )
  #v(0.35cm)
  #grid(
    columns: (auto, auto, auto, auto, auto),
    gutter: 0.25cm,
    align: horizon,
    text(size: 13pt, fill: muted)[例如：],
    pill[NaCl],
    text(size: 18pt, fill: teal)[→],
    pill(fill: coral)[离子晶体],
    box(fill: white, stroke: 0.8pt + linegray, radius: 4pt, inset: (x: 0.3cm, y: 0.12cm), text(size: 13pt)[6 配位 · NaCl 型 · $a = 5.66$ Å]),
  )
  #v(0.3cm)
  #text(size: 13pt, fill: muted)[这条流水线故意保持“可解释”：每一步都能用一句人话说明白。]
]

#slide(title: [第 0 步：先读懂化学式], kicker: [06 · 解析])[
  #grid(
    columns: (1.08fr, 0.92fr),
    gutter: 0.8cm,
    [
      #text(size: 17pt, weight: "bold")[把 Ca(OH)#sub[2] 拆成元素计数]
      #v(0.3cm)
      #grid(
        columns: (auto, auto, auto, auto, auto, auto, auto),
        gutter: 0.18cm,
        align: horizon,
        block(fill: ink, inset: (x: 0.35cm, y: 0.22cm), radius: 5pt)[
          #text(fill: white, size: 19pt)[Ca(OH)#sub[2]]
        ],
        text(size: 20pt, fill: teal)[→],
        block(width: 2.3cm, fill: white, stroke: 0.8pt + linegray, radius: 5pt, inset: 0.2cm, align(center, {
          text(size: 19pt, weight: "bold", fill: coral)[Ca]
          text(size: 11pt, fill: muted)[1 个]
        })),
        block(width: 2.3cm, fill: white, stroke: 0.8pt + linegray, radius: 5pt, inset: 0.2cm, align(center, {
          text(size: 19pt, weight: "bold", fill: coral)[O]
          text(size: 11pt, fill: muted)[2 个]
        })),
        block(width: 2.3cm, fill: white, stroke: 0.8pt + linegray, radius: 5pt, inset: 0.2cm, align(center, {
          text(size: 19pt, weight: "bold", fill: coral)[H]
          text(size: 11pt, fill: muted)[2 个]
        })),
      )
      #v(0.35cm)
      #text(size: 15pt)[括号表示“这一组原子一起乘”：Ca(OH)#sub[2] 里 O 和 H 各有 2 个。]
      #v(0.3cm)
      #note[
        #text(weight: "bold")[写错时，指到具体位置。] 比如“第 4 个字符看不懂”，而不是只说“格式错误”。
      ]
    ],
    [
      #card(title: [解析器在做什么？], accent: coral)[
        #grid(rows: (auto, auto, auto), row-gutter: 0.25cm,
          text(size: 15pt)[#text(weight: "bold", fill: teal)[元素符号]：一个大写字母，可跟一个小写字母。],
          text(size: 15pt)[#text(weight: "bold", fill: teal)[下标]：只作用于前一个元素或右括号。],
          text(size: 15pt)[#text(weight: "bold", fill: teal)[括号]：从内向外合并，再乘以括号后的数字。],
        )
        #v(0.25cm)
        #text(size: 13pt, fill: muted)[元素是否收录、电荷是否守恒，由后端判断；前端只做能独立判断的写法检查。]
      ]
    ],
  )
]

#slide(title: [第 1 步：先判断晶体家族], kicker: [07 · 分类])[
  #text(size: 17pt)[不同家族的“游戏规则”完全不同，所以先分类。]
  #v(0.3cm)
  #grid(
    columns: (1.05fr, 0.65fr, 1.3fr),
    stroke: 0.5pt + linegray,
    inset: 0.2cm,
    block(fill: ink, inset: 0.18cm)[#text(fill: white, weight: "bold")[输入特征]],
    block(fill: ink, inset: 0.18cm)[#text(fill: white, weight: "bold")[晶体家族]],
    block(fill: ink, inset: 0.18cm)[#text(fill: white, weight: "bold")[接下来怎么算]],
    [只有一种金属元素],
    [#text(weight: "bold", fill: sky)[金属晶体]],
    [查实验结构表：fcc / bcc / hcp],
    [C / Si / Ge 单质],
    [#text(weight: "bold", fill: coral)[共价晶体]],
    [金刚石型，键长 = 2 × 共价半径],
    [金属阳离子 + 阴离子],
    [#text(weight: "bold", fill: teal)[离子晶体]],
    [半径比 / 离子个数比 / 实验数据],
    [其他：多元素、2:3 计量…],
    [#text(weight: "bold", fill: muted)[暂不支持]],
    [明确提示，不硬猜],
  )
  #v(0.35cm)
  #note[
    先问“这是哪一类”，再决定用哪套规则；这比一上来就套公式更不容易张冠李戴。
  ]
]

#slide(title: [离子晶体：把离子当成硬球], kicker: [08 · 离子晶体])[
  #grid(
    columns: (0.92fr, 1.08fr),
    gutter: 0.8cm,
    [
      #align(center)[
        #box(width: 4.1cm, height: 2.5cm, {
          place(dx: 0cm, dy: 0cm, circle(radius: 0.8cm, fill: coral))
          place(dx: 1.6cm, dy: 0cm, circle(radius: 1.25cm, fill: sky))
          place(dx: 0.4cm, dy: 0.95cm, text(fill: white, size: 11pt, weight: "bold")[Na⁺])
          place(dx: 2.5cm, dy: 0.95cm, text(fill: white, size: 11pt, weight: "bold")[Cl⁻])
        })
        #v(0.2cm)
        #text(size: 12pt, fill: muted)[正负离子相切（二维示意）]
      ]
      #v(0.3cm)
      #note[
        #text(weight: "bold")[半径比] $= r_+ / r_-$：阳离子半径 ÷ 阴离子半径。
        比值越小，阳离子周围能塞下的阴离子越少。
      ]
    ],
    [
      #card(title: [三个关键词])[
        #grid(rows: (auto, auto, auto), row-gutter: 0.28cm,
          text(size: 15pt)[#text(weight: "bold", fill: teal)[配位数]：一个离子周围最近邻异号离子的个数。],
          text(size: 15pt)[#text(weight: "bold", fill: coral)[半径比]：阳离子半径 ÷ 阴离子半径，用来猜配位数。],
          text(size: 15pt)[#text(weight: "bold", fill: sky)[硬球模型]：把离子看成不可压缩的小球，球面刚好相切。],
        )
        #v(0.25cm)
        #text(size: 13pt, fill: muted)[这就是课本里的 Pauling 半径比规则：用几何关系猜配位数。]
      ]
    ],
  )
]

#slide(title: [半径比规则：三个门槛], kicker: [09 · 半径比])[
  #grid(
    columns: (0.225fr, 0.189fr, 0.318fr, 0.268fr),
    gutter: 0.05cm,
    block(width: 100%, fill: muted, height: 1.1cm, align(center + horizon, text(fill: white, size: 12pt, weight: "bold")[< 0.225])),
    block(width: 100%, fill: coral, height: 1.1cm, align(center + horizon, text(fill: white, size: 12pt, weight: "bold")[0.225 – 0.414])),
    block(width: 100%, fill: teal, height: 1.1cm, align(center + horizon, text(fill: white, size: 12pt, weight: "bold")[0.414 – 0.732])),
    block(width: 100%, fill: sky, height: 1.1cm, align(center + horizon, text(fill: white, size: 12pt, weight: "bold")[> 0.732])),
  )
  #v(0.35cm)
  #grid(
    columns: (1fr, 1fr, 1fr, 1fr),
    gutter: 0.3cm,
    card(title: [3 配位])[
      本平台不覆盖；会返回“半径比太小”，不硬猜。
    ],
    card(title: [4 配位], accent: coral)[
      ZnS 型（闪锌矿）；例子：ZnS。
    ],
    card(title: [6 配位], accent: teal)[
      NaCl 型（岩盐）；例子：NaCl、KCl。
    ],
    card(title: [8 配位], accent: sky)[
      CsCl 型（氯化铯）；例子：CsCl。
    ],
  )
  #v(0.35cm)
  #note[
    #text(weight: "bold")[分界值是经验值。] 比值贴近 0.225 / 0.414 / 0.732 时，两种配位数都可能；结果会带
    #text(font: "Consolas", size: 13pt)[caveat_near_boundary] 提醒。
  ]
]

#slide(title: [算例：NaCl 的五个步骤], kicker: [10 · 完整算例])[
  #grid(
    columns: (1.18fr, 0.82fr),
    gutter: 0.7cm,
    [
      #grid(rows: (auto, auto, auto, auto, auto), row-gutter: 0.12cm,
        step("1", [读式], [Na : Cl = 1 : 1，两种元素。]),
        step("2", [定电荷], [Na#super[+] 与 Cl#super[−] 电荷守恒，属于离子晶体。], accent: coral),
        step("3", [查半径], [$r_+ = 102$ pm，$r_- = 181$ pm，$r_+ / r_- approx 0.564$。], accent: amber),
        step("4", [判配位], [$0.414 < 0.564 < 0.732$，所以 6 配位，选 NaCl 型。], accent: teal),
        step("5", [算棱长], [$a = 2(r_+ + r_-) = 2(102 + 181)$ pm $= 5.66$ Å。], accent: sky),
      )
    ],
    [
      #align(center)[
        #cell2d(size: 6.2cm)
        #v(0.15cm)
        #text(size: 11pt, fill: muted)[二维切片：红球 Na⁺，蓝球 Cl⁻]
      ]
      #v(0.25cm)
      #card(title: [结果], accent: coral)[
        #text(size: 26pt, weight: "bold", fill: coral)[NaCl 型]
        #v(0.05cm)
        #text(size: 14pt)[配位数 6 · $a = 5.66$ Å]
        #v(0.1cm)
        #text(size: 10.5pt, fill: muted)[课本值 5.64 Å；1 Å = 100 pm。一个晶胞含 4 个 NaCl 单元。]
      ]
    ],
  )
]

#slide(title: [当经验规则遇到实验], kicker: [11 · 规则 vs 实验])[
  #grid(
    columns: (0.72fr, 1.18fr, 1.02fr, 0.92fr),
    stroke: 0.5pt + linegray,
    inset: 0.18cm,
    block(fill: ink, inset: 0.18cm)[#text(fill: white, weight: "bold")[例子]],
    block(fill: ink, inset: 0.18cm)[#text(fill: white, weight: "bold")[经验规则会怎么说]],
    block(fill: ink, inset: 0.18cm)[#text(fill: white, weight: "bold")[实验是什么]],
    block(fill: ink, inset: 0.18cm)[#text(fill: white, weight: "bold")[平台怎么标]],
    [#text(weight: "bold")[KCl]],
    [比值约 0.76 → 8 配位（CsCl 型）],
    [NaCl 型（6 配位）],
    [#text(font: "Consolas", size: 12pt)[rule_experimental_override]],
    [#text(weight: "bold")[ZnO]],
    [比值约 0.529 → 6 配位（NaCl 型）],
    [六方纤锌矿型（4 配位）],
    [#text(font: "Consolas", size: 12pt)[rule_wurtzite_experimental]],
    [#text(weight: "bold")[ZnS]],
    [比值约 0.402 → 4 配位（立方 ZnS 型）],
    [六方纤锌矿型（4 配位）],
    [#text(font: "Consolas", size: 12pt)[rule_wurtzite_experimental]],
    [#text(weight: "bold")[MgF#sub[2] / TiO#sub[2]]],
    [1:2 → 萤石型],
    [金红石型],
    [#text(font: "Consolas", size: 12pt)[rule_rutile_experimental]],
  )
  #v(0.3cm)
  #note[
    #text(weight: "bold")[这不是“作弊”，而是让模型对现实保持诚实。] 规则给出方向，实验负责校准；
    这些差异会写进结果解释，学生能看到“模型 vs 现实”。
  ]
  #v(0.2cm)
  #text(size: 12pt, fill: muted)[表中的英文是程序里的规则代号；界面上会按当前语言翻成一句人话。]
  #v(0.3cm)
  #text(size: 13pt)[教学提示：先让学生按规则算一遍，再对照实验；ZnO 就是最直观的“规则失灵”例子。]
]


#slide(title: [1:2 / 2:1：萤石型与反萤石型], kicker: [12 · 配比])[
  #grid(
    columns: (0.92fr, 1.08fr),
    gutter: 0.8cm,
    [
      #align(center)[
        #box(width: 6.4cm, height: 6.4cm, {
          let cells = range(9).map(index => {
            let row = calc.floor(index / 3)
            let col = index - row * 3
            let big = (calc.rem(row, 2) == 0 and calc.rem(col, 2) == 0) or (row == 1 and col == 1)
            align(center + horizon, if big {
              circle(radius: 0.44cm, fill: coral)
            } else {
              circle(radius: 0.22cm, fill: sky)
            })
          })
          grid(columns: (1fr,) * 3, rows: (1fr,) * 3, ..cells)
        })
        #v(0.15cm)
        #text(size: 11pt, fill: muted)[二维示意：大球 Ca#super[2+] 做骨架，小球 F#super[−] 填空隙]
      ]
      #v(0.3cm)
      #note[
        #text(weight: "bold")[配位数]：Ca#super[2+] 周围 8 个 F#super[−]；F#super[−] 周围 4 个 Ca#super[2+]。
      ]
    ],
    [
      #card(title: [萤石型 · CaF#sub[2]])[
        Ca#super[2+] 与 F#super[−] 的个数比是 1:2；大离子做面心立方骨架，小离子填四面体空隙。
      ]
      #v(0.3cm)
      #card(title: [反萤石型 · Na#sub[2]O], accent: coral)[
        反过来：O#super[2−] 做骨架，Na#super[+] 填空隙。名字里的“反”指角色互换。
      ]
      #v(0.3cm)
      #note[
        结构型由离子个数比决定；若半径比偏离 8 配位区间，会附注提醒。
      ]
    ],
  )
]

#slide(title: [金属晶体：为什么只能查表], kicker: [13 · 金属晶体])[
  #grid(
    columns: (1fr, 1fr, 1fr),
    gutter: 0.4cm,
    mini-card([面心立方 fcc], packing("fcc", size: 3.6cm)),
    mini-card([体心立方 bcc], packing("bcc", size: 3.6cm)),
    mini-card([六方最密 hcp], packing("hcp", size: 3.6cm)),
  )
  #v(0.35cm)
  #grid(
    columns: (1fr, 1fr, 1fr),
    gutter: 0.4cm,
    card(title: [Cu 铜])[面心立方：角上与每个面中心都有原子。],
    card(title: [Fe 铁], accent: coral)[体心立方：角上之外，体心还有一个原子。],
    card(title: [Mg 镁], accent: sky)[六方最密堆积：六边形一层层叠起来。],
  )
  #v(0.35cm)
  #note[
    #text(weight: "bold")[成分只告诉你有哪种原子，不告诉你怎么堆。] 同一种金属在不同温度下也可能换结构；
    所以结构型查实验表，并注明“来自实验”，不是从成分推出来的。
  ]
]

#slide(title: [共价晶体：金刚石与闪锌矿], kicker: [14 · 共价晶体])[
  #grid(
    columns: (0.9fr, 1.1fr),
    gutter: 0.8cm,
    [
      #align(center)[
        #tetra(size: 6.2cm)
        #v(0.15cm)
        #text(size: 11pt, fill: muted)[二维示意：中心原子与 4 个邻居成键]
      ]
    ],
    [
      #card(title: [C / Si / Ge → 金刚石型])[
        每个原子与 4 个邻居形成四面体网络；键长 = 2 × 共价半径。
        共价键有方向性，所以很硬，但不是“硬球相切”。
      ]
      #v(0.3cm)
      #card(title: [SiC → 闪锌矿型], accent: coral)[
        两种非金属 1:1；电负性（吸引电子的能力）小的一方占 Zn 位，另一方占 S 位。
        结构与金刚石型很像，只是两种原子交替。
      ]
      #v(0.3cm)
      #note[“金刚石型 / 闪锌矿型”是结构原型的名字，不是只有金刚石或 ZnS 才能用。]
    ],
  )
]


#section-page([PART 02 · 结果与边界], [一个诚实的教学模型], [结果怎么读，能算什么，不能算什么], accent: coral)

#slide(title: [结果怎么读：不只给一个名字], kicker: [15 · 结果解释])[
  #grid(
    columns: (0.98fr, 1.02fr),
    gutter: 0.8cm,
    [
      #card(title: [结果卡（示意）], accent: coral)[
        #text(size: 24pt, weight: "bold", fill: coral)[NaCl 型]
        #v(0.08cm)
        #text(size: 15pt)[配位数 6 · $a = 5.66$ Å]
        #v(0.2cm)
        #text(weight: "bold", fill: teal)[为什么]
        #text(size: 13pt)[半径比 0.564，落在 0.414–0.732 的 6 配位区间。]
        #v(0.15cm)
        #text(weight: "bold", fill: coral)[注意]
        #text(size: 13pt)[比值接近边界；实验支持 NaCl 型。]
      ]
    ],
    [
      #card(title: [rule：为什么])[
        #text(size: 14pt)[后端给的是 #text(font: "Consolas")[rule_radius_ratio] + 参数：
        比值 0.564、配位数 6。前端再翻成一句人话。]
      ]
      #v(0.3cm)
      #card(title: [caveats：哪里不确定], accent: coral)[
        #text(size: 14pt)[#text(font: "Consolas")[caveat_near_boundary]：贴近分界时，另一种配位数也可能。]
        #v(0.1cm)
        #text(size: 14pt)[#text(font: "Consolas")[caveat_confirmed_by_experiment]：实验支持这个结果。]
      ]
      #v(0.3cm)
      #note[同一个结果可以切换中/英文；后端只给 code + 参数，不把某一种语言的提示写死。]
    ],
  )
  #v(0.3cm)
  #text(size: 13pt, fill: muted)[结果页把“结构 / 为什么 / 注意”三块分开显示，学生不必在一大段技术文字里找重点。]
]

#slide(title: [让等待与错误都说人话], kicker: [16 · 工程保障])[
  #grid(
    columns: (1fr, 1fr),
    rows: (5.3cm, 5.3cm),
    gutter: 0.35cm,
    card(title: [异步任务], height: 100%)[
      创建后立刻返回“排队中”；后台程序计算，页面每秒更新；排队中可取消，超时不会一直转圈。
      #v(0.1cm)
      长耗时要有反馈，学生不会以为页面卡死。
    ],
    card(title: [错误可翻译], accent: coral, height: 100%)[
      后端只给错误码 + 参数；前端按当前语言翻译；没收录的错误码有兜底提示，细节进控制台。
      #v(0.1cm)
      学生只看到能读懂的句子。
    ],
    card(title: [三层追踪], accent: amber, height: 100%)[
      一次点击生成一个请求编号；三层日志用同一个编号串起来，出问题能定位。
      #v(0.1cm)
      编号会从创建任务一直带到计算层。
    ],
    card(title: [安全渲染], accent: sky, height: 100%)[
      Three.js（三维渲染库）按需加载，首页与预测页不会下载 3D 库；构建产物不含 eval / Function，
      严格的内容安全策略不需要放行 eval。
    ],
  )
  #v(0.3cm)
  #note[这些细节学生看不到，但决定了“好不好用、出问题能不能查”。]
]

#slide(title: [能算什么、不能算什么], kicker: [17 · 支持范围])[
  #grid(
    columns: (1.05fr, 0.95fr),
    gutter: 0.8cm,
    [
      #card(title: [能算], accent: teal)[
        #grid(rows: (auto, auto, auto), row-gutter: 0.25cm,
          text(size: 15pt)[#text(weight: "bold")[离子晶体]：1:1 / 1:2 / 2:1；NaCl、CsCl、ZnS、CaF#sub[2]、Na#sub[2]O、MgF#sub[2]、TiO#sub[2]],
          text(size: 15pt)[#text(weight: "bold")[金属单质]：Cu、Fe、Mg 等],
          text(size: 15pt)[#text(weight: "bold")[共价晶体]：C、Si、Ge、SiC],
        )
      ]
    ],
    [
      #card(title: [不能算], accent: coral)[
        #grid(rows: (auto, auto, auto, auto), row-gutter: 0.2cm,
          text(size: 15pt)[多元素：CaCO#sub[3]],
          text(size: 15pt)[2:3 计量：Al#sub[2]O#sub[3]],
          text(size: 15pt)[3 配位结构],
          text(size: 15pt)[复杂多晶型 / 能量排序；只画单个晶胞，不做周期扩展],
        )
      ]
    ],
  )
  #v(0.35cm)
  #note[
    遇到不支持，平台会明确说明，不会硬给一个看似合理的答案。
    这不是完整的 CSP 求解器，而是把课本规则变成可交互实验的教学模型。
  ]
]


#slide(title: [把算法变成一堂科学课], kicker: [18 · 教学价值])[
  #grid(
    columns: (1.05fr, 0.95fr),
    gutter: 0.8cm,
    [
      #grid(
        columns: (1fr, 1fr),
        rows: (auto, auto),
        gutter: 0.35cm,
        card(title: [① 模型])[把课本规则写成可执行的判断。],
        card(title: [② 预测], accent: coral)[输入化学式，得到结构、解释和附注。],
        card(title: [③ 对照], accent: amber)[把结果和实验数据比一比。],
        card(title: [④ 说明边界], accent: sky)[哪里可能错，为什么。],
      )
    ],
    [
      #text(size: 19pt, weight: "bold")[学生看到的不是“标准答案”，而是一条可检验的推理链。]
      #v(0.3cm)
      #grid(
        rows: (auto, auto, auto),
        row-gutter: 0.2cm,
        text(size: 16pt)[“为什么这么判？”],
        text(size: 16pt)[“哪里可能错？”],
        text(size: 16pt)[“实验支持吗？”],
      )
      #v(0.3cm)
      #note[#text(weight: "bold")[规则给出方向，实验负责校准，附注负责诚实。]]
    ],
  )
  #v(0.4cm)
  #note[从“记住结构”到“理解模型”：这才是平台真正想教的东西。]
]

#slide(title: [试一试，然后继续问], kicker: [19 · 结尾])[
  #grid(
    columns: (1.1fr, 0.9fr),
    gutter: 0.8cm,
    [
      #text(size: 20pt, weight: "bold")[打开平台，输入一个化学式：]
      #v(0.35cm)
      #grid(
        columns: (auto, auto, auto, auto, auto, auto),
        gutter: 0.25cm,
        pill[NaCl],
        pill(fill: coral)[CaF#sub[2]],
        pill(fill: sky)[Cu],
        pill(fill: amber)[C],
        pill(fill: teal)[SiC],
        pill(fill: muted)[ZnO],
      )
      #v(0.45cm)
      #text(size: 15pt)[看看它给出什么结构、为什么、哪里不确定。]
      #v(0.45cm)
      #note[
        下一个问题：如果两种结构能量差不多，平台会怎么告诉你？——
        找 #text(font: "Consolas")[caveat_polymorph]（多晶型：同一成分的另一种结构）。
      ]
      #v(0.7cm)
      #text(size: 12pt, fill: muted)[演示地址：http://localhost:5173/（本地启动）]
    ],
    [
      #align(center)[
        #cell2d(size: 6.4cm)
        #v(0.15cm)
        #text(size: 11pt, fill: muted)[NaCl：最常见的入门例子]
      ]
    ],
  )
  #v(0.35cm)
  #grid(
    columns: (1fr, auto),
    align: horizon,
    text(size: 12pt, fill: muted)[ycsp · 从化学式到晶体城市],
    text(size: 16pt, weight: "bold", fill: coral)[谢谢],
  )
]

