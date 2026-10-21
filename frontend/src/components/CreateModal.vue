<template>
  <div v-if="show" class="modal-backdrop">
    <div class="modal-card">
      <div class="modal-header">新建压缩包</div>
      <div class="modal-body">
        <div class="form-group">
          <label>选择待压缩文件 / 文件夹：</label>
          <div style="display: flex; gap: 8px;">
            <button class="btn" @click="selectFiles" style="flex: 1;">
              浏览选择文件...
            </button>
          </div>
          <span class="file-hint">
            {{ stagedFiles.length > 0 ? `已选择 ${stagedFiles.length} 个文件/文件夹` : '尚未选择文件' }}
          </span>
        </div>

        <div class="form-group">
          <label>压缩格式：</label>
          <div class="radio-group">
            <label><input type="radio" v-model="format" value="zip" /> ZIP (推荐)</label>
            <label><input type="radio" v-model="format" value="7z" /> 7-Zip</label>
            <label><input type="radio" v-model="format" value="tar.gz" /> TAR.GZ</label>
          </div>
        </div>

        <div class="form-group">
          <label>访问密码 (AES-256 加密，可选)：</label>
          <input
            type="password"
            v-model="password"
            class="form-input"
            placeholder="留空表示不设置密码"
          />
        </div>

        <div class="checkbox-row">
          <input type="checkbox" id="cleanMeta" v-model="cleanMetadata" />
          <label for="cleanMeta">净化 macOS 冗余垃圾文件 (.DS_Store / __MACOSX，推荐发给 Windows)</label>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn" @click="$emit('close')">取消</button>
        <button class="btn btn-primary" @click="submit">开始压缩</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { CompressOptions } from '../types/models'

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submit', opts: CompressOptions): void
}>()

const stagedFiles = ref<string[]>([])
const format = ref<'zip' | '7z' | 'tar.gz'>('zip')
const password = ref('')
const cleanMetadata = ref(true)

async function selectFiles() {
  if (window.go?.main?.App) {
    try {
      const files = await window.go.main.App.SelectFilesToCompress()
      if (files && files.length > 0) {
        stagedFiles.value = files
      }
    } catch (e) {
      console.error(e)
    }
  }
}

async function submit() {
  if (stagedFiles.value.length === 0) {
    alert('请先选择要压缩的文件')
    return
  }

  let ext = '.zip'
  if (format.value === '7z') ext = '.7z'
  if (format.value === 'tar.gz') ext = '.tar.gz'

  let destPath = ''
  if (window.go?.main?.App) {
    destPath = await window.go.main.App.SaveArchiveDialog('archive' + ext)
  }
  if (!destPath) return

  emit('submit', {
    SourcePaths: stagedFiles.value,
    DestPath: destPath,
    Format: format.value,
    Password: password.value,
    CleanMetadata: cleanMetadata.value,
  })
}
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
  width: 440px;
  box-shadow: var(--shadow-popup);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-header {
  font-size: 15px;
  font-weight: 600;
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-size: 13px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.file-hint {
  font-size: 11px;
  color: var(--text-secondary);
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
  gap: 14px;
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 6px;
}

.btn {
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-subtle);
  background: var(--bg-surface);
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.btn:hover {
  background: var(--accent-subtle);
  border-color: var(--accent);
}

.btn-primary {
  background: var(--accent);
  color: #fff;
  border: none;
}

.btn-primary:hover {
  background: var(--accent-hover);
}
</style>
