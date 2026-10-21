<template>
  <div class="breadcrumbs-bar">
    <div class="crumb-node" @click="$emit('navigate', '')">
      <span>📦</span> 根目录
    </div>
    <template v-for="(crumb, idx) in crumbs" :key="crumb.path">
      <span class="crumb-separator">></span>
      <div class="crumb-node" @click="$emit('navigate', crumb.path)">
        {{ crumb.name }}
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  currentDirectory: string
}>()

defineEmits<{
  (e: 'navigate', path: string): void
}>()

const crumbs = computed(() => {
  if (!props.currentDirectory) return []
  const clean = props.currentDirectory.replace(/^\/+|\/+$/g, '')
  if (!clean) return []

  const parts = clean.split('/')
  const list: { name: string; path: string }[] = []
  let accumulated = ''

  parts.forEach((p) => {
    accumulated += (accumulated ? '/' : '') + p
    list.push({
      name: p,
      path: accumulated,
    })
  })
  return list
})
</script>

<style scoped>
.breadcrumbs-bar {
  height: 34px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-surface);
}

.crumb-node {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  transition: color 0.15s;
}

.crumb-node:hover {
  color: var(--accent);
}

.crumb-separator {
  color: var(--text-secondary);
  opacity: 0.6;
}
</style>
