<template>
  <div
    v-if="show"
    ref="menuRef"
    class="context-menu"
    :style="computedStyle"
    @click.stop
    @contextmenu.prevent
  >
    <!-- Item-specific actions (when clicked on a file/folder) -->
    <template v-if="item">
      <div class="menu-header">
        <span class="header-icon">{{ item.is_dir ? '📁' : '📄' }}</span>
        <span class="header-name" :title="item.name">{{ item.name }}</span>
      </div>
      <div class="menu-divider"></div>

      <div class="menu-item" @click="handleAction('quickLook')">
        <div class="item-left">
          <Eye :size="14" class="menu-icon" />
          <span>快速预览 (QuickLook)</span>
        </div>
        <span class="shortcut">空格</span>
      </div>

      <div class="menu-item" v-if="!item.is_dir" @click="handleAction('openSystem')">
        <div class="item-left">
          <ExternalLink :size="14" class="menu-icon" />
          <span>以系统默认应用打开</span>
        </div>
      </div>

      <div class="menu-item" @click="handleAction('extractSelected')">
        <div class="item-left">
          <Download :size="14" class="menu-icon" />
          <span>解压所选项目...</span>
        </div>
      </div>

      <div class="menu-divider"></div>

      <div class="menu-item" @click="handleAction('copyPath')">
        <div class="item-left">
          <Copy :size="14" class="menu-icon" />
          <span>拷贝相对路径</span>
        </div>
      </div>

      <div class="menu-item" @click="handleAction('rename')">
        <div class="item-left">
          <Edit3 :size="14" class="menu-icon" />
          <span>重命名...</span>
        </div>
      </div>

      <div class="menu-divider"></div>

      <div class="menu-item danger" @click="handleAction('delete')">
        <div class="item-left">
          <Trash2 :size="14" class="menu-icon" />
          <span>从压缩包删除</span>
        </div>
      </div>
    </template>

    <!-- General actions (when clicked on empty space) -->
    <template v-else>
      <div class="menu-item" @click="handleAction('openArchive')">
        <div class="item-left">
          <FolderOpen :size="14" class="menu-icon" />
          <span>打开其他压缩包...</span>
        </div>
      </div>

      <div class="menu-item" @click="handleAction('appendFile')">
        <div class="item-left">
          <Plus :size="14" class="menu-icon" />
          <span>向此压缩包追加新文件...</span>
        </div>
      </div>

      <div class="menu-item" @click="handleAction('extractAll')">
        <div class="item-left">
          <Download :size="14" class="menu-icon" />
          <span>全部解压至...</span>
        </div>
      </div>

      <div class="menu-divider"></div>

      <div class="menu-item" @click="handleAction('refresh')">
        <div class="item-left">
          <RefreshCw :size="14" class="menu-icon" />
          <span>重新载入</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import {
  Eye,
  ExternalLink,
  Download,
  Copy,
  Edit3,
  Trash2,
  FolderOpen,
  Plus,
  RefreshCw,
} from 'lucide-vue-next'
import type { FileEntry } from '../types/models'

const props = defineProps<{
  show: boolean
  x: number
  y: number
  item: FileEntry | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'quickLook', item: FileEntry): void
  (e: 'openSystem', item: FileEntry): void
  (e: 'extractSelected', item: FileEntry): void
  (e: 'copyPath', path: string): void
  (e: 'rename', item: FileEntry): void
  (e: 'delete', item: FileEntry): void
  (e: 'openArchive'): void
  (e: 'appendFile'): void
  (e: 'extractAll'): void
  (e: 'refresh'): void
}>()

const menuRef = ref<HTMLDivElement | null>(null)
const posX = ref(0)
const posY = ref(0)

watch(
  () => [props.show, props.x, props.y],
  () => {
    if (props.show) {
      calculatePosition()
    }
  }
)

function calculatePosition() {
  const menuWidth = 230
  const menuHeight = props.item ? 240 : 150

  let left = props.x
  let top = props.y

  if (left + menuWidth > window.innerWidth) {
    left = window.innerWidth - menuWidth - 10
  }
  if (top + menuHeight > window.innerHeight) {
    top = window.innerHeight - menuHeight - 10
  }

  posX.value = Math.max(8, left)
  posY.value = Math.max(8, top)
}

const computedStyle = computed(() => ({
  left: `${posX.value}px`,
  top: `${posY.value}px`,
}))

function handleAction(actionName: string) {
  switch (actionName) {
    case 'quickLook':
      if (props.item) emit('quickLook', props.item)
      break
    case 'openSystem':
      if (props.item) emit('openSystem', props.item)
      break
    case 'extractSelected':
      if (props.item) emit('extractSelected', props.item)
      break
    case 'copyPath':
      if (props.item) emit('copyPath', props.item.path)
      break
    case 'rename':
      if (props.item) emit('rename', props.item)
      break
    case 'delete':
      if (props.item) emit('delete', props.item)
      break
    case 'openArchive':
      emit('openArchive')
      break
    case 'appendFile':
      emit('appendFile')
      break
    case 'extractAll':
      emit('extractAll')
      break
    case 'refresh':
      emit('refresh')
      break
  }
  emit('close')
}

function handleGlobalClick(e: MouseEvent) {
  if (props.show) {
    emit('close')
  }
}

onMounted(() => {
  window.addEventListener('click', handleGlobalClick)
})

onUnmounted(() => {
  window.removeEventListener('click', handleGlobalClick)
})
</script>

<style scoped>
.context-menu {
  position: fixed;
  z-index: 9999;
  width: 220px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2), 0 0 1px rgba(0, 0, 0, 0.2);
  padding: 6px;
  backdrop-filter: blur(25px);
  -webkit-backdrop-filter: blur(25px);
  animation: menuFadeIn 0.1s ease-out;
}

@keyframes menuFadeIn {
  from {
    opacity: 0;
    transform: scale(0.96);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.menu-header {
  padding: 4px 8px 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-secondary);
}

.header-name {
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.menu-divider {
  height: 1px;
  background: var(--border-subtle);
  margin: 4px 0;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 8px;
  border-radius: 5px;
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.1s ease;
}

.item-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.menu-icon {
  color: var(--text-secondary);
}

.menu-item:hover {
  background: var(--accent);
  color: #fff;
}

.menu-item:hover .menu-icon,
.menu-item:hover .shortcut {
  color: #fff;
}

.menu-item.danger:hover {
  background: var(--danger);
  color: #fff;
}

.shortcut {
  font-size: 11px;
  color: var(--text-secondary);
  background: rgba(128, 128, 128, 0.15);
  padding: 1px 4px;
  border-radius: 3px;
}
</style>
