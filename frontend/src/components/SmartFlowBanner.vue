<template>
  <div v-if="flow" class="flow-banner">
    <div class="flow-left">
      <span class="flow-icon">{{ flow.icon || '🚀' }}</span>
      <div class="flow-texts">
        <span class="flow-title">{{ flow.title }}</span>
        <span class="flow-subtitle">{{ flow.subtitle }}</span>
      </div>
    </div>
    <div class="flow-actions">
      <button class="btn btn-primary" @click="$emit('execute', flow)">
        立即执行
      </button>
      <button class="btn-close" @click="$emit('dismiss')" title="忽略此推荐">
        ✕
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FlowAction } from '../types/models'

defineProps<{
  flow: FlowAction | null
}>()

defineEmits<{
  (e: 'execute', flow: FlowAction): void
  (e: 'dismiss'): void
}>()
</script>

<style scoped>
.flow-banner {
  background: var(--accent-subtle);
  border-bottom: 1px solid var(--border-subtle);
  padding: 8px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  animation: slideDown 0.2s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.flow-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.flow-icon {
  font-size: 18px;
}

.flow-texts {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.flow-title {
  font-weight: 600;
  color: var(--text-primary);
}

.flow-subtitle {
  color: var(--text-secondary);
  font-size: 12px;
}

.flow-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn {
  padding: 5px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  background: var(--accent);
  color: white;
  border: none;
  transition: all 0.15s;
}

.btn:hover {
  background: var(--accent-hover);
  transform: translateY(-1px);
}

.btn-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.15s;
}

.btn-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}
</style>
