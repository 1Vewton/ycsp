import { describe, expect, it } from 'vitest'

import type { StructureResult } from '../src/api/tasks'
import { buildUnitCellCif } from '../src/structureCif'
import { rockSaltResult } from './fixtures'

describe('buildUnitCellCif', () => {
  it('写出 P1 晶胞参数，数值保留 4 位小数', () => {
    const cif = buildUnitCellCif(rockSaltResult)

    expect(cif).toContain('data_ycsp')
    expect(cif).toContain('_cell_length_a 5.6600')
    expect(cif).toContain('_cell_length_b 5.6600')
    expect(cif).toContain('_cell_length_c 5.6600')
    expect(cif).toContain('_cell_angle_alpha 90.0000')
    expect(cif).toContain("_symmetry_space_group_name_H-M 'P 1'")
    expect(cif).toContain('_symmetry_Int_Tables_number 1')
  })

  it('每个原子一行：唯一标签 + 元素符号 + 分数坐标', () => {
    const lines = buildUnitCellCif(rockSaltResult).split('\n')

    expect(lines).toContain('Na1 Na 0.0000 0.0000 0.0000')
    expect(lines).toContain('Cl2 Cl 0.5000 0.0000 0.0000')
  })

  it('坐标超过 4 位小数时四舍五入，不丢原子', () => {
    const shifted: StructureResult = {
      ...rockSaltResult,
      atoms: [{ symbol: 'C', fractional_coordinates: [0.123456, 0.987654, 0.5] }],
    }

    const atomLines = buildUnitCellCif(shifted)
      .split('\n')
      .filter((line) => line.startsWith('C1 '))

    expect(atomLines).toEqual(['C1 C 0.1235 0.9877 0.5000'])
  })
})
