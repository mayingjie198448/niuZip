# niuZip 详细测试方案 (Test Plan)

## 1. 测试策略与原则

本项目遵循 **测试驱动开发 (TDD)** 与自动化覆盖原则。核心层（VFS、Cleaner、Engine、In-Place Editor）单元测试覆盖率目标：**> 90%**。

---

## 2. 单元测试规划矩阵

| 测试模块 | 测试文件 | 核心测试点 |
| :--- | :--- | :--- |
| **Cleaner (清洗器)** | `core/cleaner/cleaner_test.go` | 1. 过滤 `.DS_Store`、`__MACOSX/`、`._*` 前缀文件<br>2. Windows 非法字符（`\/:*?"<>|`）正确转义<br>3. 相对路径安全性（防止 `../` 路径穿越漏洞） |
| **VFS (虚拟树与搜索)** | `core/vfs/tree_test.go` | 1. 扁平路径转换为完整树状层次结构<br>2. 深度嵌套目录的正确聚合与大小统计<br>3. 大小写不敏感的模糊搜索匹配<br>4. 中文及特殊字符（空格、Emoji、日韩文）兼容性 |
| **ZipEngine (纯Go原生ZIP)** | `core/engine/zip_test.go` | 1. 免解压极速读取目录条目与 Central Directory<br>2. 单文件按需流式提取到临时目录<br>3. 全量解压完整性比对（CRC32 / MD5）<br>4. 新建标准 ZIP 压缩包<br>5. 事务性就地删除文件（验证原包未解压仅重构）<br>6. 事务性就地重命名文件 |
| **EngineBroker (路由调度器)** | `core/engine/broker_test.go` | 1. 按文件后缀（`.zip`, `.tar.gz`, `.7z`）正确路由<br>2. 未知后缀通过 Magic Bytes 嗅探识别<br>3. 缺少对应引擎时的友好错误降级提示 |
| **In-Place Editor (事务修改)** | `core/engine/editor_test.go` | 1. 原文件原子替换前失败回滚测试<br>2. 临时文件命名隔离与生命周期测试 |

---

## 3. 边界条件与安全测试用例

1. **Zip Slip 安全漏洞防御**：
   * 输入包含恶意相对路径 `../../../../etc/passwd` 的畸形压缩包，验证解压时是否强制拦截报错。
2. **超长路径与特殊字符**：
   * 测试路径超过 260 字符（Windows MAX_PATH 边界）；
   * 测试包含特殊 Unicode 字符、空格、引号的文件名。
3. **空包与单文件包**：
   * 0 字节文件与空压缩包的处理健壮性。
4. **加密包识别**：
   * 识别是否包含加密标志位，未提供密码时提前返回清晰的 `ErrPasswordRequired`。

---

## 4. 性能基准测试规划 (Benchmarks)

* `BenchmarkVFSBuildTree_10KEntries`：构建 10,000 个节点的虚拟文件树耗时（目标：< 50ms）。
* `BenchmarkSearch_10KEntries`：在一万个条目中执行模糊搜索（目标：< 10ms）。
* `BenchmarkZipRawCopy`：就地修改 1GB ZIP 包中 1 个小文件（直接 Raw 拷贝无压缩开销，耗时 < 1s）。
