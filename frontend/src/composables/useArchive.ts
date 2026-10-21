import { ref, computed } from 'vue'
import type { ArchiveSummary, FileEntry, TreeNode, FlowAction, CompressOptions } from '../types/models'

export function useArchive() {
  const summary = ref<ArchiveSummary | null>(null)
  const currentDirectory = ref<string>('')
  const selectedItem = ref<FileEntry | null>(null)
  const searchQuery = ref<string>('')
  const searchResults = ref<FileEntry[] | null>(null)
  const activeFlow = ref<FlowAction | null>(null)
  const categoryFilter = ref<string>('all')
  const viewMode = ref<'list' | 'grid'>('list')

  const progressState = ref({
    show: false,
    title: '处理中...',
    percent: 0,
    file: '',
  })

  // Format bytes into readable format
  function formatBytes(bytes?: number): string {
    if (!bytes || bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  // Find node by path in tree
  function findNodeByPath(node: TreeNode, targetPath: string): TreeNode | null {
    const cleanTarget = targetPath.replace(/^\/+|\/+$/g, '')
    if (cleanTarget === '') return node

    const parts = cleanTarget.split('/')
    let current: TreeNode | undefined = node

    for (const part of parts) {
      if (!current || !current.children) return null
      current = current.children.find((c) => c.name === part)
    }
    return current || null
  }

  // Current items to display in table/grid
  const displayedItems = computed<FileEntry[]>(() => {
    if (searchResults.value !== null) {
      return filterByCategory(searchResults.value)
    }
    if (!summary.value || !summary.value.tree) return []

    const node = findNodeByPath(summary.value.tree, currentDirectory.value)
    if (!node || !node.children) return []

    const items: FileEntry[] = node.children.map((c) => ({
      path: c.path,
      name: c.name,
      is_dir: c.is_dir,
      size: c.size,
      packed_size: c.packed_size,
      mod_time: c.mod_time,
      is_encrypted: c.is_encrypted,
      crc32: 0,
    }))

    return filterByCategory(items)
  })

  function filterByCategory(items: FileEntry[]): FileEntry[] {
    if (categoryFilter.value === 'all') return items
    return items.filter((item) => {
      if (item.is_dir) return true
      const ext = item.name.split('.').pop()?.toLowerCase() || ''
      if (categoryFilter.value === 'images') {
        return ['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp'].includes(ext)
      }
      if (categoryFilter.value === 'docs') {
        return ['pdf', 'doc', 'docx', 'txt', 'md', 'rtf'].includes(ext)
      }
      if (categoryFilter.value === 'code') {
        return ['go', 'js', 'ts', 'py', 'json', 'html', 'css', 'c', 'cpp', 'rs', 'sh'].includes(ext)
      }
      return true
    })
  }

  function setArchive(data: ArchiveSummary) {
    summary.value = data
    currentDirectory.value = ''
    selectedItem.value = null
    searchQuery.value = ''
    searchResults.value = null
    if (data.flows && data.flows.length > 0) {
      activeFlow.value = data.flows[0]
    } else {
      activeFlow.value = null
    }
  }

  async function openArchive(path?: string) {
    if (!window.go?.main?.App) return

    let targetPath = path
    if (!targetPath) {
      targetPath = await window.go.main.App.SelectArchiveDialog()
    }
    if (!targetPath) return

    try {
      showProgress('正在解析归档目录...', 0, targetPath)
      const res = await window.go.main.App.OpenArchive(targetPath)
      hideProgress()
      setArchive(res)
    } catch (err: any) {
      hideProgress()
      alert('打开归档失败: ' + err)
    }
  }

  function navigateTo(dirPath: string) {
    currentDirectory.value = dirPath
    selectedItem.value = null
    searchResults.value = null
    searchQuery.value = ''
  }

  async function quickLook(path?: string) {
    const target = path || selectedItem.value?.path
    if (!target || !window.go?.main?.App) return
    try {
      await window.go.main.App.QuickLook(target)
    } catch (e: any) {
      console.error('QuickLook error:', e)
    }
  }

  async function openInSystem(path?: string) {
    const target = path || selectedItem.value?.path
    if (!target || !window.go?.main?.App) return
    try {
      await window.go.main.App.OpenInSystem(target)
    } catch (e: any) {
      console.error('OpenInSystem error:', e)
    }
  }

  async function extractAll() {
    if (!summary.value || !window.go?.main?.App) {
      alert('请先打开一个压缩包')
      return
    }
    try {
      const dest = await window.go.main.App.SelectDestDirDialog()
      if (!dest) return

      showProgress('正在解压所有文件...', 0, '')
      await window.go.main.App.Extract({
        archive_path: summary.value.path,
        dest_dir: dest,
        selected_files: [],
      })
      hideProgress()
      alert('全部解压完成！文件已存入: ' + dest)
    } catch (err: any) {
      hideProgress()
      alert('解压失败: ' + err)
    }
  }

  async function extractSelected(item?: FileEntry) {
    const target = item || selectedItem.value
    if (!target || !summary.value || !window.go?.main?.App) {
      alert('请先选择要解压的项目')
      return
    }
    try {
      const dest = await window.go.main.App.SelectDestDirDialog()
      if (!dest) return

      showProgress(`正在解压所选项目: ${target.name}...`, 0, target.path)
      await window.go.main.App.Extract({
        archive_path: summary.value.path,
        dest_dir: dest,
        selected_files: [target.path],
      })
      hideProgress()
      alert(`所选项目已解压至: ${dest}`)
    } catch (err: any) {
      hideProgress()
      alert('解压所选项目失败: ' + err)
    }
  }

  async function deleteSelected(item?: FileEntry) {
    const target = item || selectedItem.value
    if (!target || !summary.value || !window.go?.main?.App) {
      alert('请选择要删除的文件')
      return
    }
    if (!confirm(`确定要从当前压缩包中永久删除 "${target.name}" 吗？此操作为事务性原地修改。`)) {
      return
    }

    try {
      showProgress('正在从归档中删除条目...', 50, target.name)
      await window.go.main.App.DeleteEntries([target.path])
      hideProgress()
      selectedItem.value = null
      await refreshArchive()
    } catch (err: any) {
      hideProgress()
      alert('删除失败: ' + err)
    }
  }

  async function renameEntry(oldPath: string, newPath: string) {
    if (!summary.value || !window.go?.main?.App) return
    try {
      showProgress('正在重命名条目...', 50, newPath)
      await window.go.main.App.RenameEntry(oldPath, newPath)
      hideProgress()
      await refreshArchive(newPath)
    } catch (err: any) {
      hideProgress()
      alert('重命名失败: ' + err)
    }
  }

  async function appendFilesToArchive() {
    if (!summary.value || !window.go?.main?.App) {
      alert('请先打开一个压缩包')
      return
    }
    try {
      const files = await window.go.main.App.SelectFilesToCompress()
      if (!files || files.length === 0) return

      showProgress(`正在向归档追加 ${files.length} 个文件...`, 0, '')
      let lastTarget = ''
      for (let i = 0; i < files.length; i++) {
        const f = files[i]
        const baseName = f.split(/[\/\\]/).pop() || f
        const targetEntry = currentDirectory.value
          ? `${currentDirectory.value}/${baseName}`
          : baseName
        lastTarget = targetEntry
        await window.go.main.App.AppendEntry(f, targetEntry)
        showProgress(`正在向归档追加 ${files.length} 个文件...`, Math.round(((i + 1) / files.length) * 100), baseName)
      }
      hideProgress()
      await refreshArchive(lastTarget)
    } catch (err: any) {
      hideProgress()
      alert('追加文件失败: ' + err)
    }
  }

  async function appendDroppedFiles(filePaths: string[]) {
    if (!summary.value || !window.go?.main?.App || filePaths.length === 0) return
    try {
      showProgress(`正在向当前目录追加 ${filePaths.length} 个文件...`, 0, '')
      let lastTarget = ''
      for (let i = 0; i < filePaths.length; i++) {
        const f = filePaths[i]
        const baseName = f.split(/[\/\\]/).pop() || f
        const targetEntry = currentDirectory.value
          ? `${currentDirectory.value}/${baseName}`
          : baseName
        lastTarget = targetEntry
        await window.go.main.App.AppendEntry(f, targetEntry)
        showProgress(`正在向当前目录追加 ${filePaths.length} 个文件...`, Math.round(((i + 1) / filePaths.length) * 100), baseName)
      }
      hideProgress()
      await refreshArchive(lastTarget)
    } catch (err: any) {
      hideProgress()
      alert('追加文件失败: ' + err)
    }
  }

  async function refreshArchive(targetSelectPath?: string) {
    if (!summary.value || !window.go?.main?.App) return
    const curPath = summary.value.path
    const savedDir = currentDirectory.value

    try {
      const res = await window.go.main.App.OpenArchive(curPath)
      summary.value = res

      // Stay in current directory if it still exists
      if (savedDir !== '') {
        const node = findNodeByPath(res.tree, savedDir)
        if (node) {
          currentDirectory.value = savedDir
        } else {
          const parts = savedDir.split('/')
          parts.pop()
          const parent = parts.join('/')
          const parentNode = findNodeByPath(res.tree, parent)
          currentDirectory.value = parentNode ? parent : ''
        }
      } else {
        currentDirectory.value = ''
      }

      // Highlight target entry or re-validate selected item
      if (targetSelectPath) {
        const activeNode = findNodeByPath(res.tree, currentDirectory.value)
        const match = activeNode?.children?.find(
          (c) => c.path === targetSelectPath || c.name === targetSelectPath.split(/[\/\\]/).pop()
        )
        if (match) {
          selectedItem.value = {
            path: match.path,
            name: match.name,
            is_dir: match.is_dir,
            size: match.size,
            packed_size: match.packed_size,
            mod_time: match.mod_time,
            crc32: match.crc32,
          }
        } else {
          selectedItem.value = null
        }
      } else if (selectedItem.value) {
        const activeNode = findNodeByPath(res.tree, currentDirectory.value)
        const stillThere = activeNode?.children?.find((c) => c.path === selectedItem.value?.path)
        if (!stillThere) {
          selectedItem.value = null
        }
      }

      if (res.flows && res.flows.length > 0) {
        activeFlow.value = res.flows[0]
      } else {
        activeFlow.value = null
      }
    } catch (err) {
      console.error('Failed to refresh archive:', err)
    }
  }

  async function copyPath(pathStr: string) {
    try {
      await navigator.clipboard.writeText(pathStr)
    } catch {
      // Fallback
    }
  }

  async function handleSearch(query: string) {
    searchQuery.value = query
    if (!query || query.trim() === '') {
      searchResults.value = null
      return
    }

    if (window.go?.main?.App) {
      try {
        const results = await window.go.main.App.SearchEntries(query)
        searchResults.value = results
      } catch (e) {
        console.error('Search error:', e)
      }
    }
  }

  async function runFlow(flow: FlowAction) {
    if (!window.go?.main?.App) return
    try {
      showProgress('正在执行: ' + flow.title, 50, '')
      await window.go.main.App.ExecuteFlow(flow)
      hideProgress()
      alert('执行完成！')
    } catch (e: any) {
      hideProgress()
      alert('执行失败: ' + e)
    }
  }

  function showProgress(title: string, percent: number, file: string) {
    progressState.value = {
      show: true,
      title,
      percent,
      file,
    }
  }

  function dismissFlow() {
    activeFlow.value = null
  }

  function hideProgress() {
    progressState.value.show = false
  }

  return {
    summary,
    currentDirectory,
    selectedItem,
    searchQuery,
    searchResults,
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
  }
}
