# niuZip

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-macOS%20(ARM64)%20%7C%20Windows%20(x64)-brightgreen.svg)]()
[![License](https://img.shields.io/badge/License-MIT-purple.svg)]()

**niuZip** 是一款专为 **macOS (Apple Silicon M芯片 / arm64)** 和 **Windows x64** 打造的现代化、极速、无门槛的归档解压缩管理工具。

采用 **Go + Wails + Vue 3** 技术栈，将现代化高颜值桌面原生设计与高性能并发引擎深度结合。无任何付费门槛或功能限制，提供全功能一体化解压缩体验。

---

## ✨ 核心特性 (Key Features)

* ⚡ **免解压秒级浏览**：针对 ZIP、7z、TAR 等主流格式，毫秒级直接读取中央目录，像普通文件夹一样即时展开、折叠、翻找。
* 🎨 **高颜值现代原生交互**：
  * **列表 / 网格双视图自由切换**：支持严谨详细的列表模式与直观的 Finder 图标网格卡片模式；
  * **深度右键上下文菜单**：右键点击条目支持「QuickLook 空格预览」、「系统打开」、「解压选中项」、「复制相对路径」、「就地重命名」、「删除文件」、「追加外部文件」；右键空白区域支持快速追加与刷新；
  * **毛玻璃视觉与分类过滤**：沉浸式现代暗黑/透光风格，按图片、文档、代码等类型一键聚合筛选；
  * **多级面包屑导航**：直观层级路径展示，支持任意层级一键跳转与拖拽交互。
* 🔍 **全局模糊秒搜**：支持在超大归档内跨层级按文件名或路径即时搜索（10,000 条目仅需 0.33ms）。
* 👁 **按需临时预览 (QuickLook)**：无需解压整个压缩包，按空格键或双击即可单文件流式解压并唤起 macOS QuickLook 或 Windows 默认查看器。
* 🚀 **Smart Flow 智能快捷操作**：自动感知归档内容并推荐下一步操作（例如：检测到 `.app` 推荐一键部署至 `/Applications`；检测到字体 `.ttf/.otf` 推荐一键安装至系统字体库）。
* ✏️ **压缩包在线就地修改**：
  * 直接拖拽或右键菜单追加文件到现有压缩包；
  * 包内文件一键删除与重命名；
  * 采用**底层事务性原子替换**，零损坏原包风险。
* 🛡 **跨平台元数据净化**：
  * 向 Windows 用户发送前自动清洗 `.DS_Store`、`__MACOSX/` 等 macOS 冗余垃圾文件；
  * 自动修正 Windows 不支持的特殊字符文件名；
  * 严格防御 Zip Slip 路径穿越提权漏洞。
* 🔒 **安全加密与分卷**：
  * 支持 AES-256 高强度密码保护（支持文件名清单加密）；
  * 支持自定义分卷大小（如 100MB、1GB）切分与合并。
* 🪶 **超轻量体积 & 零外部依赖**：
  * 内置官方原生 7-Zip CLI（macOS ARM64 及 Windows x64）；
  * 运行不依赖额外运行库，打包单文件体积仅 6~7MB。

---

## 📦 下载与安装指南 (Installation Guide)

前往 [GitHub Releases 最新发布页](https://github.com/mayingjie198448/niuZip/releases/latest) 下载对应平台的安装包：

### 🍎 macOS 用户 (Apple Silicon M系列芯片)
1. 下载 **`niuZip-arm64.dmg`**；
2. 双击打开镜像，将 **niuZip** 拖入 **Applications** 文件夹即可完成安装；
3. **⚠️ 首次打开系统安全拦截说明（Gatekeeper）**：
   因个人开源项目未购买 Apple 付费开发者代码签名证书，首次打开时系统可能会提示*「已损坏，无法打开」*或*「无法验证开发者」*，这是 macOS 针对第三方软件的常规安全机制：
   * **方式一（界面操作）**：打开 macOS「系统设置」→「隐私与安全性」，滑至底部找到提示并点击 **「仍要打开」**；
   * **方式二（终端一键放行）**：打开「终端」App 执行以下命令清除隔离属性：
     ```bash
     xattr -cr /Applications/niuZip.app
     ```

### 🪟 Windows 用户 (x64)
1. 推荐下载 **`niuZip-Setup-x64.exe`**（安装版：支持自定义安装路径、自动注册「用 niuZip 打开」鼠标右键菜单与控制面板卸载），或下载 **`niuZip-windows-amd64.exe`**（免安装绿色便携版）；
2. 双击运行安装向导，按提示完成安装；
3. **⚠️ 首次打开系统安全拦截说明（SmartScreen）**：
   若 Windows Defender SmartScreen 弹出蓝色提示窗口*「Windows 已保护你的电脑」*：
   * 点击提示文本中的 **「更多信息」**；
   * 随后点击右下角出现的 **「仍要运行」** 按钮即可正常启动。

---

## 🏗 技术架构

```
niuZip
├── core/
│   ├── cleaner/       // 跨平台元数据清洗器 (Windows 友好 & Zip Slip 防护)
│   ├── engine/        // 多引擎路由调度 (PureZip, PureTar, 7zz CLI)
│   │   └── bin/       // 内嵌官方原生 7-Zip CLI (darwin_arm64 / windows_amd64)
│   └── vfs/           // 虚拟文件系统树与搜索算法
├── docs/              // 详细架构规范、API 文档与测试方案
├── assets/            // 原生静态资源 (ICNS/ICO 图标、Info.plist、DMG 视觉与打包脚本)
├── frontend/          // 现代桌面前端 (Vue 3 + Vite + TypeScript + Lucide Icons)
│   ├── src/components // HeaderBar, FileExplorer, ContextMenu, RenameModal, SmartFlow...
│   └── src/composables// 响应式业务逻辑 useArchive
└── app.go             // Wails 运行时桥接 (原生弹窗、QuickLook、SmartFlow 执行器)
```

详细架构设计请参阅：[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)  
详细接口规范请参阅：[docs/API_SPEC.md](docs/API_SPEC.md)  
测试方案请参阅：[docs/TEST_PLAN.md](docs/TEST_PLAN.md)

---

## 🚀 快速上手与本地构建

### 环境要求
* Go 1.22+
* Node.js 18+

### 编译与测试
```bash
# 运行全部测试与覆盖率报告
make test

# 构建 macOS (Apple Silicon M芯片) 原生 App 捆绑包及 DMG 安装包
make app-mac
make dmg-mac

# 构建 macOS Apple Silicon 单文件 CLI 调试执行文件
make build-mac

# 交叉编译 Windows x64 原生可执行文件 (.exe)
make build-win

# 构建 Windows x64 现代化安装包 (NSIS 封装：支持自定义路径、自动注册右键、卸载自动清除)
make installer-win

# 一键清理编译产物
make clean
```

---

## ⚖️ 免责与版权声明 (Disclaimer & Copyright)

1. **学习与交流目的**：本项目仅供个人技术研究、桌面端架构探索（Go + Wails + Vue 3）与学术交流使用，旨在向开源社区展示跨平台桌面应用的现代化工程实践。
2. **原创与知识产权**：本项目核心业务逻辑、并发调度引擎与前端交互界面均为独立自主研发；内置应用图标与 DMG 引导背景均为纯代码算法原生绘制生成；所引用的第三方工具及开源库均严格遵循其原始开源许可证（详见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)）。
3. **版权争议与快速下线通道**：本项目高度尊重任何第三方的知识产权与合法权益。若任何单位、组织或个人认为本项目的任何内容、界面交互、文案或素材涉嫌侵犯其知识产权或合法权益，**请务必通过 GitHub Issues 或邮件第一时间联系作者**并附上相关权属证明材料。作者将在核实后 24 小时内无条件优先予以删除、修改或下架处理。
4. **有限责任声明**：本项目按“现状”（AS-IS）提供，不附带任何明示或暗示的担保。任何使用者因直接或间接使用本项目代码或分发包所产生的任何后果或风险，均由使用者自行承担，作者不承担任何连带法律责任。

---

## 📄 开源许可证 (License)

* 本项目遵循 [MIT License](LICENSE) 开源许可证。
* 第三方组件及开源声明（包含 7-Zip LGPL、Wails、Vue 3 等）请参阅：[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。
