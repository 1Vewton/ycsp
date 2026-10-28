<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

import type { GLViewer } from '3dmol'
import { faRotate } from '@fortawesome/free-solid-svg-icons/faRotate'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import type { StructureResult } from '../api/tasks'
import { t } from '../locales'
import { buildUnitCellCif } from '../structureCif'

// 3Dmol 体积较大（约 157 KB gzip），只在结果页真正要画图时才动态加载。
type ThreeDmolModule = typeof import('3dmol/build/3Dmol.es6-min.js')

// 球太大会糊成一团、太小又看不出堆积方式；金属没有化学键，球稍大一点。
const SPHERE_SCALE = 0.45
const METAL_SPHERE_SCALE = 0.55
const STICK_RADIUS = 0.12
const CELL_COLOR = '#0b5d8a'
const BACKGROUND_COLOR = '#ffffff'

const props = defineProps<{
  result: StructureResult
}>()

const host = ref<HTMLDivElement | null>(null)
const isPreparing = ref(true)
const isUnavailable = ref(false)
const hasFailed = ref(false)

// 3Dmol 的句柄不放进响应式对象：它内部字段很多，代理会拖慢每帧渲染。
let viewer: GLViewer | null = null
let resizeObserver: ResizeObserver | null = null

function supportsWebGl(): boolean {
  try {
    const canvas = document.createElement('canvas')
    return Boolean(canvas.getContext('webgl2') ?? canvas.getContext('webgl'))
  } catch {
    return false
  }
}

function renderStructure(module: ThreeDmolModule): void {
  if (host.value === null) {
    return
  }

  viewer = module.createViewer(host.value, { backgroundColor: BACKGROUND_COLOR })
  const model = viewer.addModel(buildUnitCellCif(props.result), 'cif')

  const isMetallic = props.result.family === 'metallic'
  viewer.setStyle(
    {},
    isMetallic
      ? { sphere: { scale: METAL_SPHERE_SCALE } }
      : { sphere: { scale: SPHERE_SCALE }, stick: { radius: STICK_RADIUS } },
  )
  // 只画晶胞棱框：默认的红/绿/蓝 a/b/c 箭头没有标签，会让学生误以为是化学键。
  viewer.addUnitCell(model, {
    box: { color: CELL_COLOR },
    astyle: { hidden: true },
    bstyle: { hidden: true },
    cstyle: { hidden: true },
  })
  viewer.zoomTo()
  viewer.render()

  resizeObserver = new ResizeObserver(() => {
    viewer?.resize()
    viewer?.render()
  })
  resizeObserver.observe(host.value)
}

function resetView(): void {
  viewer?.zoomTo()
  viewer?.render()
}

onMounted(async () => {
  if (!supportsWebGl()) {
    isUnavailable.value = true
    isPreparing.value = false
    return
  }

  try {
    const module = await import('3dmol/build/3Dmol.es6-min.js')
    renderStructure(module)
  } catch (error) {
    // 技术细节只进控制台，界面上给学生一句能照做的提示。
    console.error(error)
    hasFailed.value = true
  } finally {
    isPreparing.value = false
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  viewer?.clear()
  viewer = null
})
</script>

<template>
  <div class="structure-viewer">
    <div class="structure-stage">
      <div
        v-if="!isUnavailable && !hasFailed"
        ref="host"
        class="structure-canvas"
        role="img"
        :aria-label="t('structure_canvas_label')"
      ></div>

      <p v-if="isPreparing" class="structure-overlay">{{ t('structure_preparing') }}</p>
      <p v-else-if="isUnavailable" class="status-error structure-error">
        {{ t('structure_webgl_unavailable') }}
      </p>
      <p v-else-if="hasFailed" class="status-error structure-error">
        {{ t('structure_render_failed') }}
      </p>
    </div>

    <div v-if="!isPreparing && !isUnavailable && !hasFailed" class="structure-controls">
      <button type="button" class="secondary-button" @click="resetView">
        <FontAwesomeIcon :icon="faRotate" aria-hidden="true" />
        <span>{{ t('structure_reset_view') }}</span>
      </button>
      <p class="structure-hint">{{ t('structure_drag_hint') }}</p>
    </div>
  </div>
</template>

<style scoped>
.structure-viewer {
  margin-top: 1rem;
}

.structure-stage {
  position: relative;
}

.structure-canvas {
  width: 100%;
  height: 22rem;
  overflow: hidden;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

/* 模型还没就绪时铺满画布区域，避免先出现一块空白。 */
.structure-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0;
  color: var(--color-text-muted);
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

.structure-error {
  margin: 0;
}

.structure-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1rem;
  align-items: center;
  margin-top: 0.75rem;
}

.structure-hint {
  margin: 0;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}
</style>
