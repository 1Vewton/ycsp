<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

import { faRotate } from '@fortawesome/free-solid-svg-icons/faRotate'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import type { StructureResult } from '../api/tasks'
import { t } from '../locales'
import type { StructureRenderer } from '../structureRenderer'
import { buildStructureGeometry } from '../structureScene'

const props = defineProps<{
  result: StructureResult
}>()

const host = ref<HTMLDivElement | null>(null)
const isPreparing = ref(true)
const isUnavailable = ref(false)
const hasFailed = ref(false)

// Three.js 体积较大；只在结果页真正要画图时才动态加载。
let renderer: StructureRenderer | null = null

function supportsWebGl(): boolean {
  try {
    // Three.js 当前版本使用 WebGL2；只有 WebGL1 的设备会走“无法显示”的提示。
    return document.createElement('canvas').getContext('webgl2') !== null
  } catch {
    return false
  }
}

onMounted(async () => {
  if (!supportsWebGl()) {
    isUnavailable.value = true
    isPreparing.value = false
    return
  }

  try {
    const geometry = buildStructureGeometry(props.result)
    const { StructureRenderer: Renderer } = await import('../structureRenderer')
    if (host.value === null) {
      return
    }
    renderer = new Renderer(host.value, geometry)
  } catch (error) {
    // 技术细节只进控制台，界面上给学生一句能照做的提示。
    console.error(error)
    hasFailed.value = true
  } finally {
    isPreparing.value = false
  }
})

function resetView(): void {
  renderer?.resetView()
}

onBeforeUnmount(() => {
  renderer?.dispose()
  renderer = null
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

      <p v-if="isPreparing" class="structure-overlay" role="status">
        {{ t('structure_preparing') }}
      </p>
      <p v-else-if="isUnavailable" class="status-error structure-error" role="alert">
        {{ t('structure_webgl_unavailable') }}
      </p>
      <p v-else-if="hasFailed" class="status-error structure-error" role="alert">
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
