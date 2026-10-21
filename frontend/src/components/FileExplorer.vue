<template>
  <div
    class="explorer-container"
    @contextmenu.prevent="handleBackgroundContextMenu"
  >
    <div v-if="items.length === 0" class="drop-hero">
      <div class="drop-hero-icon">🗂️</div>
      <div class="drop-hero-title">拖拽压缩包至此处，或点击“打开”</div>
      <div class="drop-hero-desc">
        免解压秒级浏览，支持查看图片、文档、音视频及代码，支持鼠标右键快速操作
      </div>
      <button class="hero-open-btn" @click="$emit('openArchive')">
        打开压缩包...
      </button>
    </div>

    <!-- 1. List View Mode -->
    <div v-else-if="viewMode === 'list'" class="table-scroll">
      <table>
        <thead>
          <tr>
            <th style="width: 50%;">名称</th>
            <th style="width: 15%;">大小</th>
            <th style="width: 15%;">压缩后大小</th>
            <th style="width: 20%;">修改时间</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in items"
            :key="item.path"
            :class="{ selected: selectedItem?.path === item.path }"
            @click="$emit('select', item)"
            @dblclick="handleDoubleClick(item)"
            @contextmenu.prevent.stop="handleItemContextMenu($event, item)"
          >
            <td>
              <div class="file-name-cell">
                <span class="file-icon">{{ getFileIcon(item.name, item.is_dir) }}</span>
                <span class="file-text" :title="item.name">{{ item.name }}</span>
              </div>
            </td>
            <td>{{ item.is_dir ? '--' : formatBytes(item.size) }}</td>
            <td>{{ item.is_dir ? '--' : formatBytes(item.packed_size) }}</td>
            <td>{{ item.mod_time ? formatDate(item.mod_time) : '--' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 2. Grid / Icon View Mode (Modern Finder Style) -->
    <div v-else class="grid-scroll">
      <div class="icon-grid">
        <div
          v-for="item in items"
          :key="item.path"
          class="grid-item"
          :class="{ selected: selectedItem?.path === item.path }"
          @click="$emit('select', item)"
          @dblclick="handleDoubleClick(item)"
          @contextmenu.prevent.stop="handleItemContextMenu($event, item)"
        >
          <div class="grid-icon">{{ getFileIcon(item.name, item.is_dir) }}</div>
          <div class="grid-label" :title="item.name">{{ item.name }}</div>
          <div class="grid-meta">{{ item.is_dir ? '文件夹' : formatBytes(item.size) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FileEntry } from '../types/models'

const props = defineProps<{
  items: FileEntry[]
  selectedItem: FileEntry | null
  viewMode: 'list' | 'grid'
  formatBytes: (bytes?: number) => string
}>()

const emit = defineEmits<{
  (e: 'select', item: FileEntry): void
  (e: 'drillDown', folderName: string): void
  (e: 'preview', path: string): void
  (e: 'openArchive'): void
  (e: 'itemContextMenu', event: MouseEvent, item: FileEntry): void
  (e: 'bgContextMenu', event: MouseEvent): void
}>()

function handleDoubleClick(item: FileEntry) {
  if (item.is_dir) {
    emit('drillDown', item.name)
  } else {
    emit('preview', item.path)
  }
}

function handleItemContextMenu(e: MouseEvent, item: FileEntry) {
  emit('select', item)
  emit('itemContextMenu', e, item)
}

function handleBackgroundContextMenu(e: MouseEvent) {
  emit('bgContextMenu', e)
}

function getFileIcon(name: string, isDir: boolean): string {
  if (isDir) return '📁'
  const ext = name.split('.').pop()?.toLowerCase() || ''
  if (['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp'].includes(ext)) return '🖼️'
  if (['mp4', 'mov', 'avi', 'mkv'].includes(ext)) return '🎬'
  if (['mp3', 'wav', 'flac', 'aac'].includes(ext)) return '🎵'
  if (['pdf', 'doc', 'docx', 'txt', 'md', 'rtf'].includes(ext)) return '📝'
  if (['zip', '7z', 'rar', 'tar', 'gz', 'bz2', 'xz'].includes(ext)) return '📦'
  if (['go', 'js', 'ts', 'py', 'json', 'html', 'css', 'c', 'cpp', 'rs', 'sh'].includes(ext)) return '💻'
  if (['app', 'exe', 'dmg', 'iso', 'pkg'].includes(ext)) return '⚙️'
  return '📄'
}

function formatDate(isoStr: string): string {
  try {
    const d = new Date(isoStr)
    if (isNaN(d.getTime()) || d.getFullYear() <= 1980) {
      return '--'
    }
    return d.toLocaleString()
  } catch {
    return isoStr
  }
}
</script>

<style scoped>
.explorer-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
}

.table-scroll, .grid-scroll {
  flex: 1;
  overflow-y: auto;
}

/* List Table Style */
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

thead th {
  position: sticky;
  top: 0;
  background: var(--bg-surface);
  backdrop-filter: blur(10px);
  padding: 8px 16px;
  font-weight: 600;
  font-size: 12px;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
  text-align: left;
}

tbody tr {
  cursor: pointer;
  transition: background 0.1s;
}

tbody tr:hover {
  background: rgba(0, 0, 0, 0.03);
}

@media (prefers-color-scheme: dark) {
  tbody tr:hover {
    background: rgba(255, 255, 255, 0.04);
  }
}

tbody tr.selected {
  background: var(--accent-subtle);
}

td {
  padding: 7px 16px;
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-icon {
  font-size: 16px;
}

.file-text {
  font-weight: 400;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Grid / Icon View Style (Modern Finder Style) */
.icon-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(105px, 1fr));
  gap: 16px;
  padding: 20px;
}

.grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 12px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: center;
  border: 1px solid transparent;
}

.grid-item:hover {
  background: rgba(0, 0, 0, 0.04);
}

@media (prefers-color-scheme: dark) {
  .grid-item:hover {
    background: rgba(255, 255, 255, 0.05);
  }
}

.grid-item.selected {
  background: var(--accent-subtle);
  border-color: var(--accent);
}

.grid-icon {
  font-size: 40px;
  margin-bottom: 8px;
}

.grid-label {
  font-size: 12px;
  color: var(--text-primary);
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.grid-meta {
  font-size: 11px;
  color: var(--text-secondary);
  margin-top: 2px;
}

/* Empty Hero */
.drop-hero {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: var(--text-secondary);
  padding: 60px 32px;
  text-align: center;
}

.drop-hero-icon {
  font-size: 56px;
}

.drop-hero-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.drop-hero-desc {
  font-size: 13px;
  max-width: 440px;
  line-height: 1.5;
}

.hero-open-btn {
  margin-top: 6px;
  padding: 8px 18px;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

.hero-open-btn:hover {
  background: var(--accent-hover);
}
</style>
