import type { StructureResult } from './api/tasks'

// CIF 里的数值用普通小数：分数坐标无量纲，晶胞棱长单位 Å，4 位小数足够展示。
const CIF_DECIMALS = 4

/**
 * 把后端返回的“晶胞参数 + 分数坐标”拼成只含一个晶胞的 CIF（P1 对称，不做对称扩展）。
 *
 * 原子与晶胞都由数据处理层给出，前端只换一种格式交给 3Dmol 渲染，不重算结构。
 */
export function buildUnitCellCif(result: StructureResult): string {
  const { cell } = result
  const lines = [
    'data_ycsp',
    `_cell_length_a ${formatNumber(cell.edge_a_angstrom)}`,
    `_cell_length_b ${formatNumber(cell.edge_b_angstrom)}`,
    `_cell_length_c ${formatNumber(cell.edge_c_angstrom)}`,
    `_cell_angle_alpha ${formatNumber(cell.angle_alpha_degree)}`,
    `_cell_angle_beta ${formatNumber(cell.angle_beta_degree)}`,
    `_cell_angle_gamma ${formatNumber(cell.angle_gamma_degree)}`,
    "_symmetry_space_group_name_H-M 'P 1'",
    '_symmetry_Int_Tables_number 1',
    'loop_',
    '_atom_site_label',
    '_atom_site_type_symbol',
    '_atom_site_fract_x',
    '_atom_site_fract_y',
    '_atom_site_fract_z',
  ]

  result.atoms.forEach((atom, index) => {
    const [x, y, z] = atom.fractional_coordinates
    // 标签必须唯一，用元素符号 + 序号；CIF 的第 2 列才是元素类型。
    lines.push(
      `${atom.symbol}${index + 1} ${atom.symbol} ${formatNumber(x)} ${formatNumber(y)} ${formatNumber(z)}`,
    )
  })

  return lines.join('\n')
}

function formatNumber(value: number): string {
  return value.toFixed(CIF_DECIMALS)
}
