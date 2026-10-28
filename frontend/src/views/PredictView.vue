<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { faArrowLeft } from '@fortawesome/free-solid-svg-icons/faArrowLeft'
import { faPlay } from '@fortawesome/free-solid-svg-icons/faPlay'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { createTask, problemFromError, type TaskProblem } from '../api/tasks'
import FormulaInput from '../components/FormulaInput.vue'
import { validateFormulaSyntax } from '../formulaValidation'
import { t } from '../locales'

// 现成的例子按晶体类型分组：学生可以自己输入，也可以点一下填进输入框。
interface ExampleGroup {
  labelTag: string
  formulas: string[]
}

const exampleGroups: ExampleGroup[] = [
  { labelTag: 'example_group_ionic', formulas: ['NaCl', 'CsCl', 'CaF2'] },
  { labelTag: 'example_group_covalent', formulas: ['C', 'Si', 'SiC'] },
  { labelTag: 'example_group_metallic', formulas: ['Cu', 'Fe', 'Mg'] },
]

const formula = ref('')
const isSubmitting = ref(false)
const submitFailure = ref<TaskProblem | null>(null)

// 空输入不算错，只提示“怎么写”；写法校验见 formulaValidation.ts。
const formulaProblem = computed(() =>
  formula.value.trim() === '' ? null : validateFormulaSyntax(formula.value),
)

const canSubmit = computed(
  () => formula.value.trim() !== '' && formulaProblem.value === null && !isSubmitting.value,
)

const router = useRouter()

// 改过输入就把上一次的失败提示收走，别让旧错误留在页面上。
watch(formula, () => {
  submitFailure.value = null
})

function useExample(example: string): void {
  formula.value = example
}

// 创建任务后立刻跳到结果页，由结果页轮询进度（见 ResultView.vue）。
async function startPrediction(): Promise<void> {
  if (!canSubmit.value) {
    return
  }
  isSubmitting.value = true
  submitFailure.value = null
  try {
    const task = await createTask(formula.value.trim())
    await router.push({ name: 'result', params: { taskId: String(task.id) } })
  } catch (error) {
    submitFailure.value = problemFromError(error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="page">
    <h1 class="page-title">{{ t('predict_heading') }}</h1>
    <p class="page-intro">{{ t('predict_intro') }}</p>

    <form class="prediction-form" @submit.prevent="startPrediction">
      <FormulaInput v-model="formula" :problem="formulaProblem" />

      <h2 class="section-title">{{ t('predict_examples_title') }}</h2>
      <p class="examples-hint">{{ t('predict_examples_hint') }}</p>
      <div v-for="group in exampleGroups" :key="group.labelTag" class="example-group">
        <p class="example-group-label">{{ t(group.labelTag) }}</p>
        <div class="example-chips">
          <button
            v-for="example in group.formulas"
            :key="example"
            type="button"
            class="example-chip"
            @click="useExample(example)"
          >
            {{ example }}
          </button>
        </div>
      </div>

      <button type="submit" class="primary-button submit-button" :disabled="!canSubmit">
        <FontAwesomeIcon :icon="faPlay" aria-hidden="true" />
        <span>{{ isSubmitting ? t('predict_starting') : t('predict_start_button') }}</span>
      </button>
      <p v-if="submitFailure" class="status-error" role="alert">
        {{ t(submitFailure.code, submitFailure.params) }}
      </p>
    </form>

    <p class="back-link">
      <RouterLink class="link-button" to="/">
        <FontAwesomeIcon :icon="faArrowLeft" aria-hidden="true" />
        <span>{{ t('back_home') }}</span>
      </RouterLink>
    </p>
  </section>
</template>

<style scoped>
.examples-hint {
  margin: 0 0 0.75rem;
  font-size: 0.925rem;
  color: var(--color-text-muted);
}

.example-group {
  margin-bottom: 1rem;
}

.example-group-label {
  margin: 0 0 0.35rem;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.example-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.example-chip {
  padding: 0.35rem 0.8rem;
  font-family: inherit;
  font-size: 0.95rem;
  color: var(--color-primary);
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-pill);
  cursor: pointer;
}

.example-chip:hover {
  background-color: var(--color-primary-soft);
}

.submit-button {
  margin-top: 0.5rem;
}

.back-link {
  margin: 1.5rem 0 0;
}
</style>

