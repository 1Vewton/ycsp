<script setup lang="ts">
import { computed, useId } from 'vue'

import type { FormulaProblem } from '../formulaValidation'
import { t } from '../locales'

// 输入框只负责显示与交互；写法对不对由页面用 formulaValidation.ts 判断后传进来。
const props = defineProps<{
  modelValue: string
  problem: FormulaProblem | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

// 同页复用这个组件时，label/for 与 aria-describedby 仍然各指各的。
const inputId = useId()
const feedbackId = `${inputId}-feedback`

// 反馈分三档：还没写（提示怎么写）、写错（指出第几个字符）、写法正确（可以继续）。
const feedback = computed(() => {
  if (props.problem !== null) {
    return { kind: 'error', tag: props.problem.code, params: props.problem.params ?? {} }
  }
  if (props.modelValue.trim() === '') {
    return { kind: 'hint', tag: 'formula_input_hint', params: {} }
  }
  return { kind: 'success', tag: 'formula_syntax_ok', params: {} }
})

function handleInput(event: Event): void {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>

<template>
  <div class="formula-input">
    <label class="formula-label" :for="inputId">{{ t('formula_input_label') }}</label>
    <input
      :id="inputId"
      class="formula-field"
      :class="{ 'formula-field-invalid': problem !== null }"
      type="text"
      :value="modelValue"
      :placeholder="t('formula_input_placeholder')"
      :aria-invalid="problem === null ? 'false' : 'true'"
      :aria-describedby="feedbackId"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      @input="handleInput"
    />
    <p
      :id="feedbackId"
      class="formula-feedback"
      :class="`formula-feedback-${feedback.kind}`"
      aria-live="polite"
    >
      {{ t(feedback.tag, feedback.params) }}
    </p>
  </div>
</template>

<style scoped>
.formula-input {
  padding: 1.25rem;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

.formula-label {
  display: block;
  margin-bottom: 0.4rem;
  font-weight: 600;
}

.formula-field {
  width: 100%;
  padding: 0.65rem 0.8rem;
  font-family: inherit;
  font-size: 1.05rem;
  color: var(--color-text);
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-small);
}

.formula-field:focus {
  border-color: var(--color-primary);
  outline: 2px solid var(--color-primary);
  outline-offset: 1px;
}

.formula-field-invalid {
  border-color: var(--color-error-text);
}

.formula-feedback {
  margin: 0.6rem 0 0;
  font-size: 0.925rem;
}

.formula-feedback-hint {
  color: var(--color-text-muted);
}

.formula-feedback-success {
  color: var(--color-success-text);
}

.formula-feedback-error {
  color: var(--color-error-text);
}
</style>
