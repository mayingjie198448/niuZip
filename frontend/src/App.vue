<template>
  <div class="app-layout" @click="closeContextMenu">
    <!-- 1. Header Bar with View Mode Switcher -->
    <HeaderBar
      v-model="searchQuery"
      v-model:view-mode="viewMode"
      :has-update-badge="hasUpdateBadge"
      @open="openArchive()"
      @extract="extractAll"
      @append="appendFilesToArchive"
      @create="showCreateModal = true"
      @check-update="handleCheckUpdate(false)"
      @open-github="openGitHubRepo"
    />

    <!-- 2. Smart Flow Banner (Contextual Intelligence) -->
    <SmartFlowBanner :flow="activeFlow" @execute="runFlow" @dismiss="dismissFlow" />

    <!-- 3. Main Workspace Area -->
    <div class="main-body">
      <!-- Left Sidebar: Archive Stats & Category Filters -->
      <SidebarSummary
        :summary="summary"
        :current-filter="categoryFilter"
        :format-bytes="formatBytes"
        @filter="categoryFilter = $event"
      />

      <!-- Right Explorer: Breadcrumbs & Table/Grid View -->
      <div class="workspace">
        <BreadcrumbNav
          :current-directory="currentDirectory"
          @navigate="navigateTo"
        />

        <FileExplorer
          :items="displayedItems"
          :selected-item="selectedItem"
          :view-mode="viewMode"
          :format-bytes="formatBytes"
          @select="selectedItem = $event"
          @drill-down="handleDrillDown"
          @preview="quickLook"
          @open-archive="openArchive()"
          @item-context-menu="handleItemContextMenu"
          @bg-context-menu="handleBgContextMenu"
        />
      </div>
    </div>

    <!-- 4. Bottom Action & Status Bar -->
    <footer>
      <div class="status-left">
        <span>{{ statusText }}</span>
      </div>
      <div class="status-right" v-if="selectedItem">
        <span class="selected-badge">
          已选: {{ selectedItem.name }} ({{ selectedItem.is_dir ? '目录' : formatBytes(selectedItem.size) }})
        </span>
        <button class="footer-btn" @click="quickLook(selectedItem.path)" title="快速预览 (空格)">
          预览
        </button>
        <button class="footer-btn" @click="extractSelected(selectedItem)" title="解压此项目">
          解压
        </button>
        <button class="footer-btn" @click="openRenameModal(selectedItem)" title="重命名">
          重命名
        </button>
        <button class="footer-btn danger" @click="deleteSelected(selectedItem)" title="从压缩包删除">
          删除
        </button>
      </div>
      <div class="status-right" v-else>
        <span class="os-hint">macOS Apple Silicon & Windows x64</span>
      </div>
    </footer>

    <!-- 5. Right Click Context Menu -->
    <ContextMenu
      :show="contextMenuState.show"
      :x="contextMenuState.x"
      :y="contextMenuState.y"
      :item="contextMenuState.item"
      @close="closeContextMenu"
      @quick-look="quickLook($event.path)"
      @open-system="openInSystem($event.path)"
      @extract-selected="extractSelected($event)"
      @copy-path="copyPath($event)"
      @rename="openRenameModal($event)"
      @delete="deleteSelected($event)"
      @open-archive="openArchive()"
      @append-file="appendFilesToArchive"
      @extract-all="extractAll"
      @refresh="refreshArchive()"
    />

    <!-- 6. Rename Modal -->
    <RenameModal
      :show="renameModalState.show"
      :item="renameModalState.item"
      @close="renameModalState.show = false"
      @submit="handleRenameSubmit"
    />

    <!-- 7. Create Archive Modal -->
    <CreateModal
      ref="createModalRef"
      :show="showCreateModal"
      :initial-files="createModalInitialPaths"
      @close="closeCreateModal"
      @submit="handleCreateArchive"
    />

    <!-- 8. Real-time Progress Modal -->
    <ProgressModal
      :show="progressState.show"
      :title="progressState.title"
      :percent="progressState.percent"
      :file="progressState.file"
    />

    <!-- 9. Online Software Update Modal -->
    <UpdateModal
      :show="updateState.show"
      :checking="updateState.checking"
      :downloading="updateState.downloading"
      :download-percent="updateState.downloadPercent"
      :update-info="updateState.info"
      :error-message="updateState.errorMessage"
      @close="updateState.show = false"
      @install="handleStartInstallUpdate"
      @retry="handleCheckUpdate(false)"
      @open-release="openReleaseInBrowser"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted } from 'vue'
import { useArchive } from './composables/useArchive'
import HeaderBar from './components/HeaderBar.vue'
import SmartFlowBanner from './components/SmartFlowBanner.vue'
import SidebarSummary from './components/SidebarSummary.vue'
import BreadcrumbNav from './components/BreadcrumbNav.vue'
import FileExplorer from './components/FileExplorer.vue'
import ContextMenu from './components/ContextMenu.vue'
import RenameModal from './components/RenameModal.vue'
import CreateModal from './components/CreateModal.vue'
import ProgressModal from './components/ProgressModal.vue'
import UpdateModal from './components/UpdateModal.vue'
import type { FileEntry, CompressOptions } from './types/models'

const {
  summary,
  currentDirectory,
  selectedItem,
  searchQuery,
  activeFlow,
  categoryFilter,
  viewMode,
  progressState,
  displayedItems,
  formatBytes,
  setArchive,
  openArchive,
  navigateTo,
  quickLook,
  openInSystem,
  extractAll,
  extractSelected,
  deleteSelected,
  renameEntry,
  appendFilesToArchive,
  appendDroppedFiles,
  refreshArchive,
  copyPath,
  handleSearch,
  runFlow,
  dismissFlow,
  showProgress,
  hideProgress,
} = useArchive()

const showCreateModal = ref(false)
const createModalRef = ref<any>(null)
const createModalInitialPaths = ref<string[]>([])

function closeCreateModal() {
  showCreateModal.value = false
  createModalInitialPaths.value = []
}

// Software Update State
const updateState = reactive({
  show: false,
  checking: false,
  downloading: false,
  downloadPercent: 0,
  info: null as any,
  errorMessage: '',
})

const hasUpdateBadge = ref(false)

async function handleCheckUpdate(silent = false) {
  if (!window.go?.main?.App) {
    if (!silent) {
      updateState.show = true
      updateState.errorMessage = '未检测到应用后端桥接环境'
    }
    return
  }
  if (!silent) {
    hasUpdateBadge.value = false
    updateState.show = true
    if (!updateState.info) {
      updateState.checking = true
      updateState.downloading = false
      updateState.downloadPercent = 0
      updateState.errorMessage = ''
    }
  }
  try {
    const info = await window.go.main.App.CheckForUpdate()
    updateState.info = info
    updateState.checking = false
    updateState.errorMessage = ''
    if (info && info.has_update) {
      hasUpdateBadge.value = true
    }
    if (silent) {
      // 静默自动检测发现新版本时不弹窗，仅保持红点高亮提示用户
      updateState.show = false
    }
  } catch (err: any) {
    updateState.checking = false
    if (!silent) {
      updateState.show = true
      updateState.errorMessage = String(err?.message || err)
    }
  }
}

function openGitHubRepo() {
  const url = 'https://github.com/mayingjie198448/niuZip'
  if (window.go?.main?.App) {
    window.go.main.App.OpenInSystem(url)
  } else {
    window.open(url, '_blank')
  }
}

function openReleaseInBrowser() {
  const url = updateState.info?.release_url || 'https://github.com/mayingjie198448/niuZip/releases/latest'
  if (window.go?.main?.App) {
    window.go.main.App.OpenInSystem(url)
  } else {
    window.open(url, '_blank')
  }
}

async function handleStartInstallUpdate() {
  if (!updateState.info?.download_url || !window.go?.main?.App) return
  updateState.downloading = true
  updateState.downloadPercent = 0
  try {
    await window.go.main.App.DownloadAndInstallUpdate(updateState.info.download_url)
  } catch (err: any) {
    updateState.downloading = false
    alert('下载或启动安装包失败: ' + err)
  }
}

// Context Menu State
const contextMenuState = reactive({
  show: false,
  x: 0,
  y: 0,
  item: null as FileEntry | null,
})

function handleItemContextMenu(e: MouseEvent, item: FileEntry) {
  contextMenuState.x = e.clientX
  contextMenuState.y = e.clientY
  contextMenuState.item = item
  contextMenuState.show = true
}

function handleBgContextMenu(e: MouseEvent) {
  contextMenuState.x = e.clientX
  contextMenuState.y = e.clientY
  contextMenuState.item = null
  contextMenuState.show = true
}

function closeContextMenu() {
  contextMenuState.show = false
}

// Rename Modal State
const renameModalState = reactive({
  show: false,
  item: null as FileEntry | null,
})

function openRenameModal(item: FileEntry) {
  renameModalState.item = item
  renameModalState.show = true
}

async function handleRenameSubmit(oldPath: string, newPath: string) {
  renameModalState.show = false
  await renameEntry(oldPath, newPath)
}

const statusText = computed(() => {
  if (!summary.value) return '就绪 - 请拖拽或打开一个压缩包'
  return `当前目录共 ${displayedItems.value.length} 个条目 | 全包总计 ${summary.value.total_files} 个文件, ${formatBytes(summary.value.total_uncompressed_size)}`
})

// Search watcher
watch(searchQuery, (val) => {
  handleSearch(val)
})

function handleDrillDown(folderName: string) {
  const next = currentDirectory.value
    ? `${currentDirectory.value}/${folderName}`
    : folderName
  navigateTo(next)
}

async function handleCreateArchive(opts: CompressOptions) {
  closeCreateModal()
  if (!window.go?.main?.App) return

  showProgress('正在创建压缩包...', 0, opts.DestPath)
  try {
    await window.go.main.App.Compress(opts)
    hideProgress()
    alert('创建压缩包成功！')
    await openArchive(opts.DestPath)
  } catch (err: any) {
    hideProgress()
    alert('创建失败: ' + err)
  }
}

// Keyboard shortcuts
function handleKeyDown(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
    return
  }
  if (e.code === 'Escape') {
    closeContextMenu()
    renameModalState.show = false
    closeCreateModal()
    return
  }
  if (e.code === 'Space' && selectedItem.value && !selectedItem.value.is_dir) {
    e.preventDefault()
    quickLook()
  } else if (e.code === 'Enter' && selectedItem.value) {
    e.preventDefault()
    if (selectedItem.value.is_dir) {
      handleDrillDown(selectedItem.value.name)
    } else {
      quickLook()
    }
  } else if (e.code === 'Backspace' && currentDirectory.value) {
    e.preventDefault()
    const parts = currentDirectory.value.split('/')
    parts.pop()
    navigateTo(parts.join('/'))
  }
}

onMounted(async () => {
  window.addEventListener('keydown', handleKeyDown)

  // Check if initial archive or initial paths to compress were passed via CLI/context-menu
  if (window.go?.main?.App) {
    try {
      const initialArchive = await window.go.main.App.GetInitialArchive?.()
      if (initialArchive) {
        await openArchive(initialArchive)
      } else {
        const initialCreatePaths = await window.go.main.App.GetInitialCreatePaths?.()
        if (initialCreatePaths && initialCreatePaths.length > 0) {
          createModalInitialPaths.value = initialCreatePaths
          showCreateModal.value = true
        }
      }
    } catch (e) {
      console.error('Failed to get initial launch args:', e)
    }
  }

  if (window.runtime) {
    window.runtime.EventsOn('archive-opened', (data) => {
      setArchive(data)
    })

    window.runtime.EventsOn('progress', (data) => {
      if (data.percent !== undefined) {
        progressState.value.percent = data.percent
      }
      if (data.file) {
        progressState.value.file = data.file
      }
    })

    window.runtime.EventsOn('files-dropped', async (paths: string[]) => {
      if (paths && paths.length > 0) {
        if (showCreateModal.value) {
          createModalRef.value?.addPaths(paths)
          return
        }
        if (summary.value) {
          const dirLabel = currentDirectory.value || '根目录'
          const ok = confirm(`检测到拖入 ${paths.length} 个文件，是否追加到当前目录 (${dirLabel})？\n点击"确定"追加到当前压缩包，点击"取消"创建新压缩包。`)
          if (ok) {
            await appendDroppedFiles(paths)
          } else {
            createModalInitialPaths.value = paths
            showCreateModal.value = true
          }
        } else {
          const isArchive = paths.length === 1 && /\.(zip|7z|rar|tar|gz|tgz|bz2|tbz|tbz2|xz|txz|iso|cab|apk|jar)$/i.test(paths[0])
          if (isArchive) {
            await openArchive(paths[0])
            return
          }
          createModalInitialPaths.value = paths
          showCreateModal.value = true
        }
      }
    })

    window.runtime.EventsOn('archive-file-modified', async (data: { entryPath: string; tempPath: string }) => {
      const ok = confirm(`检测到包内文件 "${data.entryPath}" 已在外部编辑器中被修改并保存。\n是否立即将最新修改同步回压缩包？`)
      if (ok && window.go?.main?.App) {
        showProgress('正在同步文件修改...', 50, data.entryPath)
        await window.go.main.App.AppendEntry(data.tempPath, data.entryPath)
        hideProgress()
        await refreshArchive(data.entryPath)
      }
    })

    window.runtime.EventsOn('update-download-progress', (data: { percent: number }) => {
      updateState.downloadPercent = Math.round(data.percent)
    })
  }

  // Silent update check 3 seconds after launch
  setTimeout(() => {
    handleCheckUpdate(true)
  }, 3000)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  position: relative;
}

.main-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.workspace {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
  overflow: hidden;
}

footer {
  height: 32px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-surface);
  font-size: 11px;
  color: var(--text-secondary);
}

.status-left {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selected-badge {
  background: var(--accent-subtle);
  color: var(--accent);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.footer-btn {
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--border-subtle);
  background: var(--bg-surface);
  color: var(--text-primary);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.footer-btn:hover {
  background: var(--accent-subtle);
  color: var(--accent);
  border-color: var(--accent);
}

.footer-btn.danger:hover {
  background: var(--danger-subtle);
  color: var(--danger);
  border-color: var(--danger);
}

.os-hint {
  opacity: 0.7;
}
</style>
