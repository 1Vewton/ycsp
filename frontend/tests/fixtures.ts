// 测试共用的样例数据；不是测试文件，所以不以 .test.ts 结尾。
import type { StructureResult, TaskSnapshot } from '../src/api/tasks'

/** 结果页与解释区用的 NaCl 结构：立方晶胞 5.66 Å，两个原子。 */
export const rockSaltResult: StructureResult = {
  family: 'ionic',
  prototype: 'rock_salt',
  coordination_number: 6,
  cell: {
    edge_a_angstrom: 5.66,
    edge_b_angstrom: 5.66,
    edge_c_angstrom: 5.66,
    angle_alpha_degree: 90,
    angle_beta_degree: 90,
    angle_gamma_degree: 90,
  },
  atoms: [
    { symbol: 'Na', fractional_coordinates: [0, 0, 0] },
    { symbol: 'Cl', fractional_coordinates: [0.5, 0, 0] },
  ],
  rule: { code: 'rule_radius_ratio', params: { ratio: 0.564, coordination_number: 6 } },
  caveats: [],
}

/** 造一份任务快照：只写当前用例关心的字段，其余用默认值。 */
export function taskSnapshot(
  status: TaskSnapshot['status'],
  overrides: Partial<TaskSnapshot> = {},
): TaskSnapshot {
  return {
    id: 7,
    formula: 'NaCl',
    status,
    created_at: '2026-11-02T00:00:00Z',
    started_at: null,
    finished_at: null,
    duration_ms: null,
    ...overrides,
  }
}
