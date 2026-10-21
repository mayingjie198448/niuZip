# niuZip 前后端接口规范 (API Specification)

本规范定义了 Go 后端核心服务（`AppService`）向前端（Wails Web 视图）暴露的 RPC 方法与数据结构。

---

## 1. 数据结构 (Data Models)

### 1.1 FileEntry (归档文件条目)
```typescript
interface FileEntry {
  path: string;         // 压缩包内相对路径，统一以 "/" 分隔
  name: string;         // 文件名 (Base Name)
  is_dir: boolean;      // 是否为目录
  size: number;         // 原始解压字节大小
  packed_size: number;  // 压缩后字节大小
  mod_time: string;     // ISO-8601 时间戳
  is_encrypted: boolean;// 是否加密
  crc32: number;        // CRC32 校验码
}
```

### 1.2 TreeNode (虚拟树节点)
```typescript
interface TreeNode {
  path: string;
  name: string;
  is_dir: boolean;
  size: number;
  mod_time: string;
  is_encrypted: boolean;
  children: TreeNode[]; // 仅目录有效，按字典序排序
}
```

### 1.3 CompressOptions (创建压缩选项)
```typescript
interface CompressOptions {
  source_paths: string[];  // 待压缩的文件/文件夹本地绝对路径列表
  dest_path: string;       // 输出归档文件绝对路径
  format: "zip" | "7z" | "tar.gz"; // 格式
  password?: string;       // AES-256 加密密码（可选）
  encrypt_header?: boolean;// 是否加密文件名清单 (7z 专有)
  split_size_mb?: number;  // 分卷大小 (MB, 0 表示不分卷)
  clean_metadata?: boolean;// 是否过滤 macOS 隐藏垃圾文件 (.DS_Store / __MACOSX)
}
```

### 1.4 ExtractOptions (解压选项)
```typescript
interface ExtractOptions {
  archive_path: string;    // 归档文件绝对路径
  dest_dir: string;        // 目标解压目录
  selected_files?: string[];// 仅解压指定相对路径（空表示全量解压）
  password?: string;       // 密码
}
```

### 1.5 ProgressEvent (解压/压缩进度通知)
```typescript
interface ProgressEvent {
  task_id: string;
  percent: number;         // 0 - 100
  current_file: string;    // 当前正在处理的文件名
  stage: "scanning" | "processing" | "completed" | "error";
  error_message?: string;
}
```

---

## 2. 后端核心方法定义 (Go AppService)

### 2.1 `OpenArchive(archivePath string) (*ArchiveSummary, error)`
* **功能**：快速读取归档头部，返回文件树与统计摘要。
* **时延目标**：< 100ms（免解压直接读取 Central Directory）。

### 2.2 `SearchEntries(query string) ([]FileEntry, error)`
* **功能**：在当前打开的归档中执行全局模糊搜索。

### 2.3 `Extract(opts ExtractOptions) error`
* **功能**：解压所选文件或全量文件至指定目录，流式上报进度。

### 2.4 `PreviewEntry(archivePath, entryPath string) (tempFilePath string, err error)`
* **功能**：将单个文件流式提取到临时缓存目录，返回系统临时文件路径，供前端调用系统 QuickLook 或外部默认应用打开。

### 2.5 `Compress(opts CompressOptions) error`
* **功能**：新建归档，支持格式切换、加密、分卷与跨平台文件清洗。

### 2.6 `ModifyArchive(archivePath string, actions []ModifyAction) error`
* **功能**：事务性在线就地修改（删除条目、追加条目、重命名条目）。
