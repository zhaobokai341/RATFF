# 文件管理器右键菜单优化

## 任务概述
将文件管理器操作列中的多个按钮（下载、属性、删除、移动、复制）整合到右键上下文菜单中，解决操作按钮过多导致列宽过大的问题。

## 完成内容

### 问题分析
原始实现中，文件管理器表格的"操作"列包含5个按钮：
- 下载 (Download)
- 属性 (Properties)
- 删除 (Delete)
- 移动 (Move)
- 复制 (Copy)

这些按钮导致操作列宽度过大，影响表格整体布局和可读性。

### 解决方案
1. 将5个操作按钮整合到右键上下文菜单中
2. 保留一个"⋯"按钮作为触发器，兼容触屏设备
3. 支持右键点击文件行直接弹出菜单
4. 添加智能定位，避免菜单超出屏幕边界

### 前端修改

#### 1. **templates/index.html** - UI组件
- 移除操作列中的5个独立按钮
- 替换为单个"⋯"按钮，点击触发上下文菜单
- 在文件行添加 `@contextmenu.prevent` 事件监听
- 新增右键上下文菜单组件，包含：
  - 下载（蓝色，↓ 图标）
  - 属性（靛蓝色，ℹ 图标）
  - 分隔线
  - 删除（红色，✕ 图标）
  - 移动（黄色，↔ 图标）
  - 复制（绿色，⧉ 图标）
- 菜单使用固定定位，支持点击外部区域关闭

#### 2. **static/js/app.js** - 前端逻辑
新增响应式状态：
- `fmContextMenuVisible` - 菜单可见性
- `fmContextMenuX` - 菜单X坐标
- `fmContextMenuY` - 菜单Y坐标
- `fmContextMenuFile` - 当前选中的文件

新增函数：
- `showContextMenu(f, event)` - 显示上下文菜单，包含边界检测逻辑
- `hideContextMenu()` - 隐藏上下文菜单
- `contextMenuDownload()` - 从菜单触发下载
- `contextMenuProperties()` - 从菜单触发属性查看
- `contextMenuDelete()` - 从菜单触发删除
- `contextMenuMove()` - 从菜单触发移动
- `contextMenuCopy()` - 从菜单触发复制

### 技术细节

#### 边界检测逻辑
```javascript
var menuWidth = 160;
var menuHeight = 200;
var windowWidth = window.innerWidth;
var windowHeight = window.innerHeight;
if (x + menuWidth > windowWidth) {
    x = windowWidth - menuWidth - 10;
}
if (y + menuHeight > windowHeight) {
    y = windowHeight - menuHeight - 10;
}
fmContextMenuX.value = Math.max(10, x);
fmContextMenuY.value = Math.max(10, y);
```

确保菜单始终在可视区域内显示，避免被屏幕边缘裁剪。

#### 事件处理
- 右键点击文件行：`@contextmenu.prevent="showContextMenu(f, $event)"`
- 点击"⋯"按钮：`@click="showContextMenu(f, $event)"`
- 点击菜单外部：`@click="hideContextMenu"`
- 菜单内部点击：`@click.stop` 阻止事件冒泡

### 用户体验改进
1. **更简洁的界面**：操作列宽度大幅减小，表格更紧凑
2. **更快的操作**：右键直接弹出菜单，减少点击次数
3. **触屏兼容**：保留"⋯"按钮，触屏设备可点击触发
4. **视觉一致性**：菜单颜色与操作按钮颜色保持一致
5. **智能定位**：自动调整位置，避免菜单超出屏幕

### 测试验证
- ✅ Go编译通过：`go build -o /tmp/server_web_test ./server_web/`
- ✅ golangci-lint检查通过：`golangci-lint run ./server_web/...`
- ✅ JavaScript语法检查通过：`node -c server_web/static/js/app.js`

### 影响范围
- 仅修改前端UI，不涉及后端逻辑
- 复用现有的文件操作函数（downloadFile, showProperties, showDeleteFile等）
- 无需修改i18n语言文件（复用现有标签）