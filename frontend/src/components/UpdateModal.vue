<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-card">
      <!-- Header -->
      <div class="modal-header">
        <div class="header-left">
          <Sparkles class="header-icon" :size="20" />
          <h3>软件更新</h3>
        </div>
        <button class="btn-icon" @click="$emit('close')" :disabled="downloading">✕</button>
      </div>

      <!-- Body -->
      <div class="modal-body">
        <!-- Checking State -->
        <div v-if="checking" class="state-center">
          <div class="spinner"></div>
          <p class="state-text">正在检查 GitHub 最新版本...</p>
        </div>

        <!-- Already Latest State -->
        <div v-else-if="updateInfo && !updateInfo.has_update" class="state-center">
          <div class="check-badge">✓</div>
          <h4>当前已是最新版本</h4>
          <p class="version-sub">niuZip v{{ updateInfo.current_version }} 运行良好，暂无更新。</p>
        </div>

        <!-- Has Update State -->
        <div v-else-if="updateInfo && updateInfo.has_update" class="update-content">
          <div class="version-banner">
            <span class="new-tag">NEW</span>
            <div class="version-info">
              <span class="version-title">发现新版本 v{{ updateInfo.latest_version }}</span>
              <span class="version-compare">当前版本: v{{ updateInfo.current_version }}</span>
            </div>
            <span class="file-size" v-if="updateInfo.asset_size">{{ formatSize(updateInfo.asset_size) }}</span>
          </div>

          <div class="release-notes-box">
            <div class="notes-header">更新日志：</div>
            <pre class="notes-content">{{ updateInfo.release_notes || '包含性能优化与已知问题修复。' }}</pre>
          </div>

          <!-- Download Progress -->
          <div v-if="downloading" class="progress-section">
            <div class="progress-bar-bg">
              <div class="progress-bar-fill" :style="{ width: `${downloadPercent}%` }"></div>
            </div>
            <div class="progress-labels">
              <span>正在下载更新安装包...</span>
              <span class="percent-num">{{ downloadPercent }}%</span>
            </div>
          </div>
        </div>

        <!-- Error / Null fallback -->
        <div v-else class="state-center">
          <p class="state-text">未能获取版本信息，请检查网络连接。</p>
        </div>
      </div>

      <!-- Footer Buttons -->
      <div class="modal-footer">
        <template v-if="updateInfo && updateInfo.has_update">
          <button class="btn btn-secondary" @click="$emit('close')" :disabled="downloading">
            稍后再说
          </button>
          <button class="btn btn-primary" @click="$emit('install')" :disabled="downloading">
            {{ downloading ? '正在下载...' : '立即下载并更新' }}
          </button>
        </template>
        <template v-else>
          <button class="btn btn-primary" @click="$emit('close')">
            确定
          </button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Sparkles } from 'lucide-vue-next'

export interface UpdateInfoModel {
  has_update: boolean
  current_version: string
  latest_version: string
  release_title: string
  release_notes: string
  download_url: string
  asset_size: number
  asset_name: string
  release_url: string
}

defineProps<{
  show: boolean
  checking: boolean
  downloading: boolean
  downloadPercent: number
  updateInfo: UpdateInfoModel | null
}>()

defineEmits<{
  (e: 'close'): void
  (e: 'install'): void
}>()

function formatSize(bytes?: number): string {
  if (!bytes) return ''
  const mb = bytes / (1024 * 1024)
  return mb.toFixed(1) + ' MB'
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  animation: fadeIn 0.15s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-card {
  width: 480px;
  max-width: 90vw;
  background: var(--bg-card, #1e293b);
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  border-radius: 14px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-icon {
  color: #38bdf8;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #f8fafc);
}

.btn-icon {
  background: transparent;
  border: none;
  color: var(--text-secondary, #94a3b8);
  cursor: pointer;
  font-size: 16px;
  border-radius: 6px;
  padding: 4px 8px;
  transition: all 0.15s;
}

.btn-icon:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.modal-body {
  padding: 24px 20px;
  min-height: 140px;
}

.state-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 20px 0;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(56, 189, 248, 0.2);
  border-top-color: #38bdf8;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.state-text {
  color: var(--text-secondary, #94a3b8);
  font-size: 14px;
}

.check-badge {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
  font-size: 24px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 14px;
}

.state-center h4 {
  margin: 0 0 6px 0;
  font-size: 16px;
  color: #f8fafc;
}

.version-sub {
  margin: 0;
  font-size: 13px;
  color: var(--text-secondary, #94a3b8);
}

.version-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: rgba(56, 189, 248, 0.1);
  border: 1px solid rgba(56, 189, 248, 0.25);
  border-radius: 10px;
  margin-bottom: 16px;
}

.new-tag {
  background: #0284c7;
  color: white;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  letter-spacing: 0.5px;
}

.version-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.version-title {
  font-size: 14px;
  font-weight: 600;
  color: #f8fafc;
}

.version-compare {
  font-size: 12px;
  color: var(--text-secondary, #94a3b8);
}

.file-size {
  font-size: 12px;
  color: #38bdf8;
  font-weight: 500;
}

.release-notes-box {
  background: rgba(0, 0, 0, 0.25);
  border-radius: 8px;
  padding: 12px 14px;
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.notes-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary, #94a3b8);
  margin-bottom: 8px;
}

.notes-content {
  margin: 0;
  font-family: inherit;
  font-size: 13px;
  color: #cbd5e1;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 140px;
  overflow-y: auto;
  line-height: 1.5;
}

.progress-section {
  margin-top: 16px;
}

.progress-bar-bg {
  height: 8px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 6px;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #38bdf8, #2563eb);
  transition: width 0.15s ease-out;
}

.progress-labels {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-secondary, #94a3b8);
}

.percent-num {
  font-weight: 600;
  color: #38bdf8;
}

.modal-footer {
  padding: 14px 20px;
  background: rgba(0, 0, 0, 0.2);
  border-top: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.08));
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.btn {
  padding: 8px 18px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.15s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary, #f8fafc);
}

.btn-secondary:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.15);
}

.btn-primary {
  background: #0284c7;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #0369a1;
}
</style>
