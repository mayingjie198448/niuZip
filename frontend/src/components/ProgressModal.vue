<template>
  <div v-if="show" class="modal-backdrop">
    <div class="modal-card">
      <div class="modal-header">{{ title }}</div>
      <div class="modal-body">
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: percent + '%' }"></div>
        </div>
        <div class="progress-info">
          <span class="file-text">{{ file || '处理中...' }}</span>
          <span class="percent-text">{{ Math.round(percent) }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  show: boolean
  title: string
  percent: number
  file: string
}>()
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius);
  width: 400px;
  box-shadow: var(--shadow-popup);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.modal-header {
  font-size: 14px;
  font-weight: 600;
}

.progress-track {
  height: 8px;
  background: var(--bg-base);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--accent);
  transition: width 0.15s ease;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: var(--text-secondary);
}

.file-text {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.percent-text {
  font-weight: 500;
  color: var(--text-primary);
}
</style>
