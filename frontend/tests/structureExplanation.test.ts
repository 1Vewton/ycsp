import { describe, expect, it } from 'vitest'

import {
  buildAngleMeasures,
  buildEdgeMeasures,
  describeCellShape,
  formatMeasure,
} from '../src/structureExplanation'
import { rockSaltResult } from './fixtures'

const cubicCell = rockSaltResult.cell

describe('describeCellShape', () => {
  it('三条棱等长且夹角都是 90° 时判为立方', () => {
    expect(describeCellShape(cubicCell)).toBe('cubic')
  })

  it('容差内的差异仍算等长、直角（避免浮点误差改口）', () => {
    expect(
      describeCellShape({ ...cubicCell, edge_b_angstrom: 5.7, angle_gamma_degree: 89.6 }),
    ).toBe('cubic')
  })

  it('底面等长、三个夹角都是 90°、但 c 不同时判为四方', () => {
    expect(describeCellShape({ ...cubicCell, edge_c_angstrom: 7.2 })).toBe('tetragonal')
  })

  it('底面等长、γ=120° 且 α=β=90° 时判为六方', () => {
    expect(
      describeCellShape({
        ...cubicCell,
        edge_c_angstrom: 5.2,
        angle_gamma_degree: 120,
      }),
    ).toBe('hexagonal')
  })

  it('其它情况归为普通晶胞', () => {
    expect(describeCellShape({ ...cubicCell, edge_b_angstrom: 5.0 })).toBe('general')
    expect(describeCellShape({ ...cubicCell, angle_gamma_degree: 100 })).toBe('general')
  })
})

describe('晶胞参数格式化', () => {
  it('去掉末尾多余的 0', () => {
    expect(formatMeasure(5.6, 2)).toBe('5.6')
    expect(formatMeasure(5.66, 2)).toBe('5.66')
    expect(formatMeasure(90, 1)).toBe('90')
    expect(formatMeasure(0, 2)).toBe('0')
  })

  it('三条棱与三个夹角都带标签', () => {
    expect(buildEdgeMeasures(cubicCell)).toEqual([
      { label: 'a', value: '5.66' },
      { label: 'b', value: '5.66' },
      { label: 'c', value: '5.66' },
    ])
    expect(buildAngleMeasures(cubicCell)).toEqual([
      { label: 'α', value: '90' },
      { label: 'β', value: '90' },
      { label: 'γ', value: '90' },
    ])
  })
})
