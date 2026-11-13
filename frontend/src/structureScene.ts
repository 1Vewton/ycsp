import type { StructureCell, StructureResult } from './api/tasks'

/** 三维空间中的一点，单位 Å。 */
export interface Point3 {
  x: number
  y: number
  z: number
}

/** 一个原子及其在晶胞中的位置（Å）。 */
export interface PlacedAtom {
  symbol: string
  position: Point3
}

/** 一根连线（化学键或晶胞棱）的两端（Å）。 */
export interface Segment {
  start: Point3
  end: Point3
}

/** 三维视图需要的几何数据；渲染器只消费这些数，不再读化学数据。 */
export interface StructureGeometry {
  atoms: PlacedAtom[]
  bonds: Segment[]
  cellEdges: Segment[]
  cellCorners: Point3[]
  center: Point3
  radius: number
  atomRadius: number
}

/** 三条棱向量，单位 Å；a 沿 x 轴，b 在 x-y 平面内。 */
export interface CellVectors {
  a: Point3
  b: Point3
  c: Point3
}

// 键长判定：两原子距离不超过最短原子间距的这个倍数，视为第一配位壳层。
const BOND_DISTANCE_TOLERANCE = 1.2
// 球半径与最短原子间距的比例；金属没有化学键，球稍大一些才能看出堆积。
const ATOM_RADIUS_RATIO = 0.32
const METALLIC_ATOM_RADIUS_RATIO = 0.45
// 只有一个原子（没有原子对）时仍要让球可见。
const MIN_ATOM_RADIUS_ANGSTROM = 0.25
// 没有原子对可比较时的间距兜底，避免半径为 0。
const FALLBACK_PAIR_DISTANCE_ANGSTROM = 1.0

const DEGREE_TO_RADIAN = Math.PI / 180

/** 把晶胞参数换成三条棱向量（Å）。 */
export function cellVectors(cell: StructureCell): CellVectors {
  const alpha = cell.angle_alpha_degree * DEGREE_TO_RADIAN
  const beta = cell.angle_beta_degree * DEGREE_TO_RADIAN
  const gamma = cell.angle_gamma_degree * DEGREE_TO_RADIAN

  const sinGamma = Math.sin(gamma)
  if (Math.abs(sinGamma) < Number.EPSILON) {
    // 后端已经保证 0° < γ < 180°；这里只是防止坏数据算出 Infinity。
    throw new Error('晶胞夹角 γ 不能是 0° 或 180°。')
  }

  const a = { x: cell.edge_a_angstrom, y: 0, z: 0 }
  const b = {
    x: cell.edge_b_angstrom * Math.cos(gamma),
    y: cell.edge_b_angstrom * sinGamma,
    z: 0,
  }
  const cX = cell.edge_c_angstrom * Math.cos(beta)
  const cY =
    (cell.edge_c_angstrom * (Math.cos(alpha) - Math.cos(beta) * Math.cos(gamma))) / sinGamma
  const cZ = Math.sqrt(
    Math.max(0, cell.edge_c_angstrom * cell.edge_c_angstrom - cX * cX - cY * cY),
  )
  return { a, b, c: { x: cX, y: cY, z: cZ } }
}

/** 分数坐标 → 直角坐标（Å）。 */
export function fractionalToCartesian(
  coordinates: readonly [number, number, number],
  vectors: CellVectors,
): Point3 {
  const [x, y, z] = coordinates
  return {
    x: vectors.a.x * x + vectors.b.x * y + vectors.c.x * z,
    y: vectors.a.y * x + vectors.b.y * y + vectors.c.y * z,
    z: vectors.a.z * x + vectors.b.z * y + vectors.c.z * z,
  }
}

/** 把后端的结构结果换成渲染器需要的原子、键与晶胞线框。 */
export function buildStructureGeometry(result: StructureResult): StructureGeometry {
  const vectors = cellVectors(result.cell)
  const atoms = result.atoms.map((atom) => ({
    symbol: atom.symbol,
    position: fractionalToCartesian(atom.fractional_coordinates, vectors),
  }))
  const cellCorners = buildCellCorners(vectors)
  const center = {
    x: (vectors.a.x + vectors.b.x + vectors.c.x) / 2,
    y: (vectors.a.y + vectors.b.y + vectors.c.y) / 2,
    z: (vectors.a.z + vectors.b.z + vectors.c.z) / 2,
  }
  const shortestDistance = shortestPairDistance(atoms, result.family)
  const atomRadius = Math.max(
    MIN_ATOM_RADIUS_ANGSTROM,
    shortestDistance *
      (result.family === 'metallic' ? METALLIC_ATOM_RADIUS_RATIO : ATOM_RADIUS_RATIO),
  )

  return {
    atoms,
    bonds: buildBonds(result.family, atoms, shortestDistance),
    cellEdges: buildCellEdges(cellCorners),
    cellCorners,
    center,
    radius: sceneRadius(center, cellCorners, atoms),
    atomRadius,
  }
}

function buildCellCorners(vectors: CellVectors): Point3[] {
  const corners: Point3[] = []
  for (const x of [0, 1]) {
    for (const y of [0, 1]) {
      for (const z of [0, 1]) {
        corners.push(fractionalToCartesian([x, y, z] as const, vectors))
      }
    }
  }
  return corners
}

// 顶点顺序与 buildCellCorners 一致：x 每 +1 跳过 4 个，y 跳过 2 个，z 跳过 1 个。
const CELL_EDGE_INDEX_PAIRS = [
  [0, 1],
  [0, 2],
  [0, 4],
  [1, 3],
  [1, 5],
  [2, 3],
  [2, 6],
  [3, 7],
  [4, 5],
  [4, 6],
  [5, 7],
  [6, 7],
] as const

function buildCellEdges(corners: Point3[]): Segment[] {
  return CELL_EDGE_INDEX_PAIRS.map(([startIndex, endIndex]) => ({
    start: corners[startIndex],
    end: corners[endIndex],
  }))
}

function shortestPairDistance(atoms: PlacedAtom[], family: string): number {
  let shortest = Number.POSITIVE_INFINITY
  for (let first = 0; first < atoms.length; first += 1) {
    for (let second = first + 1; second < atoms.length; second += 1) {
      // 离子晶体只比较不同种离子：同种离子之间不是化学键。
      if (family === 'ionic' && atoms[first].symbol === atoms[second].symbol) {
        continue
      }
      shortest = Math.min(shortest, distance(atoms[first].position, atoms[second].position))
    }
  }
  return Number.isFinite(shortest) ? shortest : FALLBACK_PAIR_DISTANCE_ANGSTROM
}

function buildBonds(family: string, atoms: PlacedAtom[], shortestDistance: number): Segment[] {
  if (family === 'metallic') {
    return []
  }

  const limit = shortestDistance * BOND_DISTANCE_TOLERANCE
  const bonds: Segment[] = []
  for (let first = 0; first < atoms.length; first += 1) {
    for (let second = first + 1; second < atoms.length; second += 1) {
      if (family === 'ionic' && atoms[first].symbol === atoms[second].symbol) {
        continue
      }
      if (distance(atoms[first].position, atoms[second].position) <= limit) {
        bonds.push({ start: atoms[first].position, end: atoms[second].position })
      }
    }
  }
  return bonds
}

function sceneRadius(center: Point3, corners: Point3[], atoms: PlacedAtom[]): number {
  let radius = 0
  for (const point of [...corners, ...atoms.map((atom) => atom.position)]) {
    radius = Math.max(radius, distance(center, point))
  }
  return radius > 0 ? radius : FALLBACK_PAIR_DISTANCE_ANGSTROM
}

function distance(first: Point3, second: Point3): number {
  const deltaX = first.x - second.x
  const deltaY = first.y - second.y
  const deltaZ = first.z - second.z
  return Math.sqrt(deltaX * deltaX + deltaY * deltaY + deltaZ * deltaZ)
}
