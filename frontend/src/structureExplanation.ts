import type { StructureCell } from './api/tasks'

// 展示精度：棱长到 0.01 Å，角度到 0.1°；末尾的 0 会去掉，避免出现“90.0°”。
const EDGE_DECIMALS = 2
const ANGLE_DECIMALS = 1

// 判断晶胞形状时的容差：数据本身是精确值，留一点余量避免浮点误差。
const EDGE_TOLERANCE_ANGSTROM = 0.05
const ANGLE_TOLERANCE_DEGREE = 0.5

const RIGHT_ANGLE_DEGREE = 90
const HEXAGONAL_GAMMA_DEGREE = 120

/** 晶胞形状：立方、六方，或无法归入这两类的普通晶胞。 */
export type CellShape = 'cubic' | 'hexagonal' | 'general'

/** 一条要展示的晶胞参数，例如 { label: 'a', value: '5.66' }。 */
export interface CellMeasure {
  label: string
  value: string
}

/** 按棱长与夹角判断晶胞形状，返回语言包里的 cell_shape_* 后缀。 */
export function describeCellShape(cell: StructureCell): CellShape {
  const equalBaseEdges = isClose(
    cell.edge_a_angstrom,
    cell.edge_b_angstrom,
    EDGE_TOLERANCE_ANGSTROM,
  )

  if (
    equalBaseEdges &&
    isClose(cell.edge_a_angstrom, cell.edge_c_angstrom, EDGE_TOLERANCE_ANGSTROM) &&
    isRightAngle(cell.angle_alpha_degree) &&
    isRightAngle(cell.angle_beta_degree) &&
    isRightAngle(cell.angle_gamma_degree)
  ) {
    return 'cubic'
  }

  if (
    equalBaseEdges &&
    isRightAngle(cell.angle_alpha_degree) &&
    isRightAngle(cell.angle_beta_degree) &&
    isClose(cell.angle_gamma_degree, HEXAGONAL_GAMMA_DEGREE, ANGLE_TOLERANCE_DEGREE)
  ) {
    return 'hexagonal'
  }

  return 'general'
}

/** 三条棱长，单位 Å。 */
export function buildEdgeMeasures(cell: StructureCell): CellMeasure[] {
  return [
    { label: 'a', value: formatMeasure(cell.edge_a_angstrom, EDGE_DECIMALS) },
    { label: 'b', value: formatMeasure(cell.edge_b_angstrom, EDGE_DECIMALS) },
    { label: 'c', value: formatMeasure(cell.edge_c_angstrom, EDGE_DECIMALS) },
  ]
}

/** 三个夹角，单位 °。 */
export function buildAngleMeasures(cell: StructureCell): CellMeasure[] {
  return [
    { label: 'α', value: formatMeasure(cell.angle_alpha_degree, ANGLE_DECIMALS) },
    { label: 'β', value: formatMeasure(cell.angle_beta_degree, ANGLE_DECIMALS) },
    { label: 'γ', value: formatMeasure(cell.angle_gamma_degree, ANGLE_DECIMALS) },
  ]
}

/** 保留指定小数位，并去掉末尾多余的 0（90.00 → 90，5.60 → 5.6）。 */
export function formatMeasure(value: number, decimals: number): string {
  return value.toFixed(decimals).replace(/\.?0+$/, '')
}

function isRightAngle(angleDegree: number): boolean {
  return isClose(angleDegree, RIGHT_ANGLE_DEGREE, ANGLE_TOLERANCE_DEGREE)
}

function isClose(left: number, right: number, tolerance: number): boolean {
  return Math.abs(left - right) <= tolerance
}
