import { describe, expect, it } from 'vitest'

import type { StructureResult } from '../src/api/tasks'
import { buildStructureGeometry, cellVectors } from '../src/structureScene'
import { rockSaltResult } from './fixtures'

describe('cellVectors', () => {
  it('立方晶胞的三条棱互相垂直', () => {
    const vectors = cellVectors(rockSaltResult.cell)

    expect(vectors.a.x).toBeCloseTo(5.66)
    expect(vectors.a.y).toBe(0)
    expect(vectors.b.x).toBeCloseTo(0)
    expect(vectors.b.y).toBeCloseTo(5.66)
    expect(vectors.c.z).toBeCloseTo(5.66)
  })

  it('六方晶胞的 b 棱落在 120° 方向', () => {
    const vectors = cellVectors({
      ...rockSaltResult.cell,
      edge_a_angstrom: 3,
      edge_b_angstrom: 3,
      edge_c_angstrom: 5,
      angle_gamma_degree: 120,
    })

    expect(vectors.b.x).toBeCloseTo(-1.5)
    expect(vectors.b.y).toBeCloseTo(2.598, 3)
    expect(vectors.c.z).toBeCloseTo(5)
  })

  it('γ 为 180° 时拒绝退化晶胞', () => {
    expect(() =>
      cellVectors({ ...rockSaltResult.cell, angle_gamma_degree: 180 }),
    ).toThrow()
  })
})

describe('buildStructureGeometry', () => {
  it('把分数坐标换成 Å，并画出顶点、棱与 NaCl 键', () => {
    const geometry = buildStructureGeometry(rockSaltResult)

    expect(geometry.atoms).toHaveLength(2)
    expect(geometry.atoms[0].position).toEqual({ x: 0, y: 0, z: 0 })
    expect(geometry.atoms[1].position.x).toBeCloseTo(2.83)
    expect(geometry.cellCorners).toHaveLength(8)
    expect(geometry.cellEdges).toHaveLength(12)
    expect(geometry.center.x).toBeCloseTo(2.83)
    expect(geometry.bonds).toHaveLength(1)
    expect(geometry.atomRadius).toBeCloseTo(2.83 * 0.32)
  })

  it('离子晶体不把同种离子之间画成键', () => {
    const result: StructureResult = {
      ...rockSaltResult,
      atoms: [
        { symbol: 'Ca', fractional_coordinates: [0, 0, 0] },
        { symbol: 'F', fractional_coordinates: [0.25, 0.25, 0.25] },
        { symbol: 'F', fractional_coordinates: [0.5, 0, 0] },
      ],
    }

    expect(buildStructureGeometry(result).bonds).toHaveLength(2)
  })

  it('金刚石型画出 C–C 键，金属晶体不画键', () => {
    const diamond: StructureResult = {
      ...rockSaltResult,
      family: 'covalent',
      prototype: 'diamond',
      atoms: [
        { symbol: 'C', fractional_coordinates: [0, 0, 0] },
        { symbol: 'C', fractional_coordinates: [0.25, 0.25, 0.25] },
      ],
    }
    expect(buildStructureGeometry(diamond).bonds).toHaveLength(1)

    const metal: StructureResult = {
      ...rockSaltResult,
      family: 'metallic',
      prototype: 'body_centered_cubic',
      atoms: [
        { symbol: 'Cu', fractional_coordinates: [0, 0, 0] },
        { symbol: 'Cu', fractional_coordinates: [0.5, 0.5, 0.5] },
      ],
    }
    const metalGeometry = buildStructureGeometry(metal)
    expect(metalGeometry.bonds).toHaveLength(0)
    expect(metalGeometry.atomRadius).toBeCloseTo(4.901 * 0.45, 3)
  })
})
