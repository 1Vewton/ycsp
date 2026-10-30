<script setup lang="ts">
import { computed } from 'vue'

import type { StructureResult } from '../api/tasks'
import { t } from '../locales'
import {
  buildAngleMeasures,
  buildEdgeMeasures,
  describeCellShape,
} from '../structureExplanation'

const props = defineProps<{
  result: StructureResult
}>()

// 只翻译已知晶体类型；后端以后多出一种类型时，至少还能显示原始值而不是空白。
const FAMILY_TAGS: Record<string, string> = {
  ionic: 'family_ionic',
  covalent: 'family_covalent',
  metallic: 'family_metallic',
}

const familyText = computed(() => {
  const tag = FAMILY_TAGS[props.result.family]
  return tag === undefined ? props.result.family : t(tag)
})

const prototypeText = computed(() => t(`prototype_${props.result.prototype}`))
const shapeText = computed(() => t(`cell_shape_${describeCellShape(props.result.cell)}`))
const edgeMeasures = computed(() => buildEdgeMeasures(props.result.cell))
const angleMeasures = computed(() => buildAngleMeasures(props.result.cell))
</script>

<template>
  <section class="explanation">
    <h2 class="section-title">{{ t('explain_title') }}</h2>

    <div class="explain-card">
      <dl class="explain-facts">
        <div class="explain-fact">
          <dt>{{ t('explain_family_label') }}</dt>
          <dd>{{ familyText }}</dd>
        </div>
        <div class="explain-fact">
          <dt>{{ t('explain_prototype_label') }}</dt>
          <dd>{{ prototypeText }}</dd>
        </div>
        <div class="explain-fact">
          <dt>{{ t('explain_coordination_label') }}</dt>
          <dd>{{ result.coordination_number }}</dd>
        </div>
      </dl>
      <p class="explain-note">{{ t('explain_coordination_note') }}</p>
    </div>

    <div class="explain-card">
      <h3 class="explain-heading">{{ t('explain_cell_title') }}</h3>
      <p class="explain-text">{{ shapeText }}</p>

      <div class="cell-group">
        <p class="cell-group-title">{{ t('explain_edge_title') }}</p>
        <ul class="cell-list">
          <li v-for="measure in edgeMeasures" :key="measure.label">
            {{ measure.label }} = {{ measure.value }} Å
          </li>
        </ul>
      </div>

      <div class="cell-group">
        <p class="cell-group-title">{{ t('explain_angle_title') }}</p>
        <ul class="cell-list">
          <li v-for="measure in angleMeasures" :key="measure.label">
            {{ measure.label }} = {{ measure.value }}°
          </li>
        </ul>
      </div>

      <p class="explain-note">{{ t('explain_unit_note') }}</p>
    </div>

    <div class="explain-card">
      <h3 class="explain-heading">{{ t('explain_rule_title') }}</h3>
      <p class="explain-text">{{ t(result.rule.code, result.rule.params) }}</p>
    </div>

    <div v-if="result.caveats.length > 0" class="explain-card">
      <h3 class="explain-heading">{{ t('explain_caveats_title') }}</h3>
      <ul class="explain-list">
        <li v-for="(caveat, index) in result.caveats" :key="index">
          {{ t(caveat.code, caveat.params) }}
        </li>
      </ul>
    </div>
  </section>
</template>

<style scoped>
.explanation {
  margin-top: 2rem;
}

.explain-card {
  margin-top: 1rem;
  padding: 1.25rem;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

.explain-facts {
  display: grid;
  gap: 0.75rem 1.5rem;
  margin: 0;
}

@media (min-width: 40rem) {
  .explain-facts {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.explain-fact dt {
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

.explain-fact dd {
  margin: 0.15rem 0 0;
  font-weight: 600;
}

.explain-heading {
  margin: 0 0 0.5rem;
  font-size: 1.05rem;
}

.explain-text {
  margin: 0;
  line-height: 1.7;
}

.explain-note {
  margin: 0.75rem 0 0;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.cell-group {
  margin-top: 1rem;
}

.cell-group-title {
  margin: 0 0 0.35rem;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.cell-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1.25rem;
  margin: 0;
  padding: 0;
  list-style: none;
  font-variant-numeric: tabular-nums;
}

.explain-list {
  margin: 0.25rem 0 0;
  padding-left: 1.2rem;
  line-height: 1.7;
}

.explain-list li + li {
  margin-top: 0.35rem;
}
</style>
