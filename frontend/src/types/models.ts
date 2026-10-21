export interface FileEntry {
  path: string
  name: string
  is_dir: boolean
  size: number
  packed_size: number
  mod_time: string
  is_encrypted: boolean
  crc32: number
}

export interface TreeNode {
  path: string
  name: string
  is_dir: boolean
  size: number
  packed_size: number
  mod_time: string
  is_encrypted: boolean
  children?: TreeNode[]
}

export interface FlowAction {
  id: string
  type: 'install_app' | 'install_font' | 'preview_media'
  title: string
  subtitle: string
  icon: string
  payload: string
}

export interface ArchiveSummary {
  path: string
  filename: string
  format: string
  total_files: number
  total_folders: number
  total_uncompressed_size: number
  total_packed_size: number
  has_encryption: boolean
  tree?: TreeNode
  flows?: FlowAction[]
}

export interface CompressOptions {
  SourcePaths: string[]
  DestPath: string
  Format: string
  Password?: string
  EncryptHeader?: boolean
  SplitSizeMB?: number
  CleanMetadata?: boolean
}

export interface ExtractOptions {
  archive_path: string
  dest_dir: string
  selected_files: string[]
  password?: string
}

export interface ProgressEventData {
  percent?: number
  file?: string
  stage?: string
}

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          OpenArchive: (path: string) => Promise<ArchiveSummary>
          SearchEntries: (query: string) => Promise<FileEntry[]>
          PreviewEntry: (entryPath: string) => Promise<string>
          QuickLook: (entryPath: string) => Promise<void>
          OpenInSystem: (entryPath: string) => Promise<void>
          Extract: (opts: ExtractOptions) => Promise<void>
          Compress: (opts: CompressOptions) => Promise<void>
          DeleteEntries: (targetPaths: string[]) => Promise<void>
          RenameEntry: (oldPath: string, newPath: string) => Promise<void>
          AppendEntry: (sourceFilePath: string, targetEntryPath: string) => Promise<void>
          DetectFlows: (summary: ArchiveSummary) => Promise<FlowAction[]>
          ExecuteFlow: (action: FlowAction) => Promise<void>
          SelectArchiveDialog: () => Promise<string>
          SelectDestDirDialog: () => Promise<string>
          SaveArchiveDialog: (defaultName: string) => Promise<string>
          SelectFilesToCompress: () => Promise<string[]>
        }
      }
    }
    runtime?: {
      EventsOn: (eventName: string, callback: (...args: any[]) => void) => void
      EventsEmit: (eventName: string, ...args: any[]) => void
    }
  }
}
