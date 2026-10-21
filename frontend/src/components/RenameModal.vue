<template>
  <div v-if="show" class="modal-backdrop" @click.self="$emit('close')">
    <div class="modal-card">
      <div class="modal-header">重命名项目</div>
      <div class="modal-body">
        <label class="label-text">输入新的文件名：</label>
        <input
          ref="inputRef"
          type="text"
          v-model="newName"
          class="form-input"
          @keydown.enter="submit"
        />
        <span class="hint-text">原路径: {{ item?.path }}</span>
      </div>
      <div class="modal-footer">
        <button class="btn" @click="$emit('close')">取消</button>
        <button class="btn btn-primary" @click="submit" :disabled="!isValid">
          确认修改
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import type { FileEntry } from '../types/models'

const props = defineProps<{
  show: boolean
  item: FileEntry | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submit', oldPath: string, newPath: string): void
}>()

const newName = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

watch(
  () => [props.show, props.item],
  () => {
    if (props.show && props.item) {
      newName.value = props.item.name
      nextTick(() => {
        if (inputRef.value) {
          inputRef.value.focus()
          // Select base name without extension
          const lastDot = newName.value.lastIndexOf('.')
          if (lastDot > 0 && !props.item?.is_dir) {
            inputRef.value.setSelectionRange(0, lastDot)
          } else {
            inputRef.value.select()
          }
        }
      })
    }
  }
)

const isValid = computed(() => {
  const trimmed = newName.value.trim()
  return trimmed.length > 0 && trimmed !== props.item?.name
})

function submit() {
  if (!isValid.value || !props.item) return
  const oldPath = props.item.path
  let newPath = ''

  // Preserve parent directory if present
  const lastSlash = oldPath.lastIndexOf('/')
  if (lastSlash >= 0) {
    newPath = oldPath.substring(0, lastSlash + 1) + newName.value.trim()
  } else {
    newPath = newName.value.trim()
  }

  if (props.item.is_dir && !newPath.endsWith('/')) {
    newPath += '/'
  }

  emit('submit', oldPath, newPath)
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
  width: 380px;
  box-shadow: var(--shadow-popup);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.modal-header {
  font-size: 15px;
  font-weight: 600;
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.label-text {
  font-size: 12px;
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

.hint-text {
  font-size: 11px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
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

.btn-primary:hover:not(:disabled) {
  background: var(--accent-hover);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
