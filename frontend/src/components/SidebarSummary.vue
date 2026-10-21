<template>
  <aside>
    <div class="sidebar-section-title">归档概况</div>
    <div class="summary-card" v-if="summary">
      <div class="stat-row">
        <span class="stat-label">文件名称:</span>
        <span class="stat-val" :title="summary.path">{{ summary.filename }}</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">归档格式:</span>
        <span class="stat-val badge">{{ summary.format.toUpperCase() }}</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">文件总数:</span>
        <span class="stat-val">{{ summary.total_files }}</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">目录总数:</span>
        <span class="stat-val">{{ summary.total_folders }}</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">解压容量:</span>
        <span class="stat-val">{{ formatBytes(summary.total_uncompressed_size) }}</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">压缩容量:</span>
        <span class="stat-val">{{ formatBytes(summary.total_packed_size) }}</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">加密保护:</span>
        <span class="stat-val">{{ summary.has_encryption ? '🔒 AES 加密' : '未加密' }}</span>
      </div>
    </div>
    <div v-else class="empty-summary">
      <p>未载入归档</p>
    </div>

    <div class="sidebar-section-title" style="margin-top: 18px;">快速筛选</div>
    <div
      class="filter-item"
      :class="{ active: currentFilter === 'all' }"
      @click="$emit('filter', 'all')"
    >
      <span>📁</span> 全部内容
    </div>
    <div
      class="filter-item"
      :class="{ active: currentFilter === 'images' }"
      @click="$emit('filter', 'images')"
    >
      <span>🖼️</span> 图片与相片
    </div>
    <div
      class="filter-item"
      :class="{ active: currentFilter === 'docs' }"
      @click="$emit('filter', 'docs')"
    >
      <span>📝</span> 文档与电子书
    </div>
    <div
      class="filter-item"
      :class="{ active: currentFilter === 'code' }"
      @click="$emit('filter', 'code')"
    >
      <span>💻</span> 代码与配置
    </div>
  </aside>
</template>

<script setup lang="ts">
import type { ArchiveSummary } from '../types/models'

defineProps<{
  summary: ArchiveSummary | null
  currentFilter: string
  formatBytes: (bytes?: number) => string
}>()

defineEmits<{
  (e: 'filter', val: string): void
}>()
</script>

<style scoped>
aside {
  width: 230px;
  border-right: 1px solid var(--border-subtle);
  background: var(--bg-sidebar);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.sidebar-section-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.summary-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.stat-label {
  color: var(--text-secondary);
}

.stat-val {
  font-weight: 500;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.badge {
  background: var(--accent-subtle);
  color: var(--accent);
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.empty-summary {
  font-size: 12px;
  color: var(--text-secondary);
}

.filter-item {
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.15s;
}

.filter-item:hover, .filter-item.active {
  background: var(--accent-subtle);
  color: var(--accent);
  font-weight: 500;
}
</style>
