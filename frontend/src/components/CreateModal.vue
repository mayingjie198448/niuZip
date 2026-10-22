<template>
  <div v-if="show" class="modal-backdrop" @click.self="$emit('close')">
    <div class="modal-card">
      <div class="modal-header">
        <div class="modal-title">
          <span>📦</span> 新建压缩包
        </div>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <div class="modal-body">
        <!-- 待压缩内容管理 -->
        <div class="form-group">
          <div class="section-title-row">
            <label>待压缩文件 / 文件夹：</label>
            <span class="file-hint" v-if="stagedItems.length > 0">
              已选 {{ stagedItems.length }} 项
            </span>
          </div>

          <div class="actions-row">
            <button class="btn btn-action" @click="selectFiles">
              📄 添加文件...
            </button>
            <button class="btn btn-action" @click="selectFolder">
              📁 添加文件夹...
            </button>
            <button
              v-if="stagedItems.length > 0"
              class="btn btn-danger-subtle"
              @click="clearAll"
              title="清空已选列表"
            >
              清空
            </button>
          </div>

          <!-- 已添加条目列表 -->
          <div v-if="stagedItems.length > 0" class="staged-list">
            <div
              v-for="(item, idx) in stagedItems"
              :key="item.path"
              class="staged-item"
              :title="item.path"
            >
              <span class="item-icon">{{ item.is_dir ? '📁' : '📄' }}</span>
              <div class="item-info">
                <span class="item-name">{{ item.name }}</span>
                <span class="item-sub">
                  {{ item.is_dir ? '文件夹' : formatSize(item.size) }} · {{ item.path }}
                </span>
              </div>
              <button
                class="btn-remove"
                @click="removeItem(idx)"
                title="移除此项"
              >
                ✕
              </button>
            </div>
          </div>

          <!-- 空状态提示 -->
          <div v-else class="empty-dropzone">
            <div class="dropzone-icon">📥</div>
            <div class="dropzone-text">尚未添加待压缩文件或文件夹</div>
            <div class="dropzone-hint">点击上方按钮选择，或直接将文件/文件夹拖入窗口</div>
          </div>
        </div>

        <!-- 压缩格式 -->
        <div class="form-group">
          <label>压缩格式：</label>
          <div class="radio-group">
            <label class="radio-label">
              <input type="radio" v-model="format" value="zip" />
              <span>ZIP (通用推荐)</span>
            </label>
            <label class="radio-label">
              <input type="radio" v-model="format" value="7z" />
              <span>7-Zip (更高压缩率)</span>
            </label>
            <label class="radio-label">
              <input type="radio" v-model="format" value="tar.gz" />
              <span>TAR.GZ</span>
            </label>
          </div>
        </div>

        <!-- 访问密码 -->
        <div class="form-group">
          <label>访问密码 (AES-256 加密，可选)：</label>
          <input
            type="password"
            v-model="password"
            class="form-input"
            placeholder="留空表示不设置密码"
          />
        </div>

        <!-- 清理垃圾元数据 -->
        <div class="checkbox-row">
          <input type="checkbox" id="cleanMeta" v-model="cleanMetadata" />
          <label for="cleanMeta">净化 macOS 冗余垃圾文件 (.DS_Store / __MACOSX，推荐发给 Windows)</label>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn" @click="$emit('close')">取消</button>
        <button
          class="btn btn-primary"
          :disabled="stagedItems.length === 0"
          @click="submit"
        >
          开始压缩 ({{ stagedItems.length }})
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { CompressOptions, PathInfo } from '../types/models'

const props = defineProps<{
  show: boolean
  initialFiles?: string[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submit', opts: CompressOptions): void
}>()

const stagedItems = ref<PathInfo[]>([])
const format = ref<'zip' | '7z' | 'tar.gz'>('zip')
const password = ref('')
const cleanMetadata = ref(true)

watch(
  () => props.show,
  async (visible) => {
    if (visible && props.initialFiles && props.initialFiles.length > 0) {
      await addPaths(props.initialFiles)
    }
  }
)

watch(
  () => props.initialFiles,
  async (newFiles) => {
    if (props.show && newFiles && newFiles.length > 0) {
      await addPaths(newFiles)
    }
  }
)

async function addPaths(paths: string[]) {
  if (!paths || paths.length === 0) return
  if (window.go?.main?.App?.GetPathInfos) {
    try {
      const infos = await window.go.main.App.GetPathInfos(paths)
      const existing = new Set(stagedItems.value.map((i) => i.path))
      for (const info of infos) {
        if (!existing.has(info.path)) {
          stagedItems.value.push(info)
          existing.add(info.path)
        }
      }
      return
    } catch (e) {
      console.error('Failed to get path infos:', e)
    }
  }

  // Fallback if backend method is unavailable
  const existing = new Set(stagedItems.value.map((i) => i.path))
  for (const p of paths) {
    if (!existing.has(p)) {
      const name = p.replace(/[\\\/]+$/, '').split(/[\\\/]/).pop() || p
      stagedItems.value.push({
        path: p,
        name: name,
        is_dir: !name.includes('.'),
        size: 0,
      })
      existing.add(p)
    }
  }
}

async function selectFiles() {
  if (window.go?.main?.App) {
    try {
      const files = await window.go.main.App.SelectFilesToCompress()
      if (files && files.length > 0) {
        await addPaths(files)
      }
    } catch (e) {
      console.error(e)
    }
  }
}

async function selectFolder() {
  if (window.go?.main?.App) {
    try {
      const folder = await window.go.main.App.SelectFolderToCompress()
      if (folder) {
        await addPaths([folder])
      }
    } catch (e) {
      console.error(e)
    }
  }
}

function removeItem(index: number) {
  stagedItems.value.splice(index, 1)
}

function clearAll() {
  stagedItems.value = []
}

function formatSize(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return parseFloat((bytes / Math.pow(1024, i)).toFixed(1)) + ' ' + units[i]
}

async function submit() {
  if (stagedItems.value.length === 0) {
    alert('请先选择要压缩的文件或文件夹')
    return
  }

  let ext = '.zip'
  if (format.value === '7z') ext = '.7z'
  if (format.value === 'tar.gz') ext = '.tar.gz'

  // Smart default filename based on the staged item
  let defaultBase = 'archive'
  if (stagedItems.value.length === 1) {
    const rawName = stagedItems.value[0].name
    defaultBase = rawName.replace(/\.[^/.]+$/, '') || rawName
  } else if (stagedItems.value.length > 1) {
    const rawName = stagedItems.value[0].name.replace(/\.[^/.]+$/, '')
    defaultBase = `${rawName}-etc`
  }

  let destPath = ''
  if (window.go?.main?.App) {
    destPath = await window.go.main.App.SaveArchiveDialog(defaultBase + ext)
  }
  if (!destPath) return

  emit('submit', {
    SourcePaths: stagedItems.value.map((i) => i.path),
    DestPath: destPath,
    Format: format.value,
    Password: password.value,
    CleanMetadata: cleanMetadata.value,
  })
}

defineExpose({
  addPaths,
})
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(5px);
}

.modal-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  width: 500px;
  max-width: 90vw;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.35);
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  animation: modalPop 0.18s ease-out;
}

@keyframes modalPop {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(-4px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 16px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}

.btn-close:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
  font-size: 13px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.form-group label {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.file-hint {
  font-size: 11px;
  color: var(--accent);
  font-weight: 500;
}

.actions-row {
  display: flex;
  gap: 8px;
}

.btn-action {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 7px 12px;
  font-size: 12px;
}

.btn-danger-subtle {
  padding: 7px 12px;
  font-size: 12px;
  color: #ff5252;
  border-color: rgba(255, 82, 82, 0.2);
}

.btn-danger-subtle:hover {
  background: rgba(255, 82, 82, 0.1);
  border-color: #ff5252;
}

.staged-list {
  max-height: 140px;
  overflow-y: auto;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-base);
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.staged-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  background: var(--bg-surface);
  transition: background 0.15s;
}

.staged-item:hover {
  background: var(--bg-hover);
}

.item-icon {
  font-size: 15px;
  flex-shrink: 0;
}

.item-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.item-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-sub {
  font-size: 10px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-remove {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  flex-shrink: 0;
}

.btn-remove:hover {
  color: #ff5252;
  background: rgba(255, 82, 82, 0.15);
}

.empty-dropzone {
  border: 1px dashed var(--border-subtle);
  border-radius: 8px;
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.015);
  text-align: center;
}

.dropzone-icon {
  font-size: 20px;
  margin-bottom: 4px;
  opacity: 0.6;
}

.dropzone-text {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.dropzone-hint {
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.7;
  margin-top: 2px;
}

.form-input {
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid var(--border-subtle);
  background: var(--bg-base);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.form-input:focus {
  border-color: var(--accent);
}

.radio-group {
  display: flex;
  gap: 16px;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 12px;
  color: var(--text-primary);
}

.radio-label input {
  cursor: pointer;
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.checkbox-row input {
  cursor: pointer;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

.btn {
  padding: 7px 16px;
  border-radius: 6px;
  border: 1px solid var(--border-subtle);
  background: var(--bg-surface);
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn:hover:not(:disabled) {
  background: var(--accent-subtle);
  border-color: var(--accent);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
  border: none;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-hover);
}
</style>
