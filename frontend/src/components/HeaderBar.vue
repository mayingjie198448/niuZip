<template>
  <header>
    <div class="brand">
      <span class="brand-icon">📦</span>
      <span class="brand-title">niuZip</span>
    </div>

    <div class="toolbar">
      <!-- Primary Actions -->
      <button class="btn" @click="$emit('open')" title="打开压缩包 (⌘O)">
        <FolderOpen :size="14" /> 打开
      </button>
      <button class="btn btn-primary" @click="$emit('extract')" title="全部解压至指定目录">
        <Download :size="14" /> 解压
      </button>
      <button class="btn" @click="$emit('append')" title="向当前压缩包追加文件">
        <FilePlus :size="14" /> 追加文件
      </button>
      <button class="btn" @click="$emit('create')" title="打包新建压缩包">
        <Plus :size="14" /> 新建压缩
      </button>

      <!-- View Mode Switcher (List vs Grid) -->
      <div class="view-mode-group">
        <button
          class="mode-btn"
          :class="{ active: viewMode === 'list' }"
          @click="$emit('update:viewMode', 'list')"
          title="列表视图"
        >
          <List :size="14" />
        </button>
        <button
          class="mode-btn"
          :class="{ active: viewMode === 'grid' }"
          @click="$emit('update:viewMode', 'grid')"
          title="图标视图"
        >
          <LayoutGrid :size="14" />
        </button>
      </div>

      <!-- Quick Search -->
      <div class="search-container">
        <Search class="search-icon" :size="13" />
        <input
          type="text"
          :value="modelValue"
          @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
          placeholder="搜索文件或路径..."
        />
      </div>

      <!-- Check Update -->
      <button class="btn btn-update" @click="$emit('check-update')" title="检查版本更新">
        <Sparkles :size="13" />
        <span>更新</span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import {
  FolderOpen,
  Download,
  Plus,
  FilePlus,
  Search,
  List,
  LayoutGrid,
  Sparkles,
} from 'lucide-vue-next'

defineProps<{
  modelValue: string
  viewMode: 'list' | 'grid'
}>()

defineEmits<{
  (e: 'update:modelValue', val: string): void
  (e: 'update:viewMode', mode: 'list' | 'grid'): void
  (e: 'open'): void
  (e: 'extract'): void
  (e: 'append'): void
  (e: 'create'): void
  (e: 'check-update'): void
}>()
</script>

<style scoped>
header {
  height: 52px;
  padding: 0 16px 0 80px; /* Space for macOS window controls */
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-surface);
  --wails-draggable: drag;
}

.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
}

.brand-icon {
  font-size: 18px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  --wails-draggable: no-drag;
}

.btn {
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid var(--border-subtle);
  background: var(--bg-surface);
  color: var(--text-primary);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s ease;
}

.btn:hover {
  background: var(--accent-subtle);
  border-color: var(--accent);
  color: var(--accent);
}

.btn-primary {
  background: var(--accent);
  color: #fff;
  border-color: transparent;
}

.btn-primary:hover {
  background: var(--accent-hover);
  color: #fff;
}

/* View Mode Segmented Control */
.view-mode-group {
  display: flex;
  background: var(--bg-base);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  padding: 2px;
  gap: 2px;
}

.mode-btn {
  padding: 4px 8px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.mode-btn.active {
  background: var(--bg-surface);
  color: var(--accent);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.search-container {
  position: relative;
  margin-left: 4px;
  --wails-draggable: no-drag;
}

.search-container input {
  background: var(--bg-base);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  padding: 6px 10px 6px 28px;
  color: var(--text-primary);
  font-size: 12px;
  width: 170px;
  outline: none;
  transition: width 0.2s ease, border-color 0.2s ease;
}

.search-container input:focus {
  width: 220px;
  border-color: var(--accent);
}

.search-icon {
  position: absolute;
  left: 8px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-secondary);
  pointer-events: none;
}

.btn-update {
  background: rgba(56, 189, 248, 0.12);
  color: #38bdf8;
  border: 1px solid rgba(56, 189, 248, 0.25);
  margin-left: 2px;
}

.btn-update:hover {
  background: rgba(56, 189, 248, 0.22);
  border-color: #38bdf8;
  transform: translateY(-1px);
}
</style>
