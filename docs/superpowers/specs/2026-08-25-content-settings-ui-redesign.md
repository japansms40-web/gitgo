# 内容设置 UI 高保真还原设计文档

**日期**: 2026-08-25  
**版本**: 1.0  
**状态**: 设计阶段

---

## 1. 需求概述

参考提供的两张 UI 设计图，对 `ContentSettingsPage.vue` 组件进行高保真还原。主要改动包括：

- UI 字段标签调整
- 新增「{AI扩写}」标签（UI 占用，暂无后端实现）
- 添加「文件库」标签页用于管理 `/docs/git/配置` 目录
- 调整下拉选择器样式和标签面板布局
- 在 `/docs/git` 目录中定义标签类型配置规则

---

## 2. 设计范围

### 前端改动
- **路径**: `/frontend/src/components/ContentSettingsPage.vue`
- **改动类型**: UI 布局、样式、组件结构调整
- **新增组件**: `FileLibraryPanel.vue`（文件库标签页）

### 后端改动
- **路径**: `/docs/git/`
- **新增文件**: `tags.json`（标签类型配置）
- **现有文件**: 保持不变（关键词、变量、文章库等）

### 范围外
- 不实现 {AI扩写} 的后端逻辑
- 不修改现有配置文件结构
- 不改动其他页面或组件

---

## 3. 详细设计

### 3.1 标签页结构（4 页）

#### 页面 1：模板（template）
**当前路径**: `/frontend/src/components/ContentSettingsPage.vue` 第 118-195 行

**改动点**:
1. 字段标签文本调整
   - 「标题模板 Title」 → 「{标题}」
   - 「正文模板 Body · 支持右侧全部标签」 → 「【内容】支持右侧所有标签」

2. 标签面板（TOKEN_COLUMNS）
   - 新增 {AI扩写} 标签
   - 重新排列三列标签的展示顺序，按参考设计排列

3. 工具栏调整
   - 在「变量设置」按钮后新增「文件库」按钮
   - 「文件库」按钮样式：白色背景 + 灰色边框（非主色调）

#### 页面 2：变量设置（bank）
**当前路径**: `/frontend/src/components/ContentSettingsPage.vue` 第 197-262 行

**改动**: 保持现有布局和功能不变

#### 页面 3：文件库（library）【新增】
**位置**: 新标签页，在「预览」标签页前

**内容**:
- 显示 `/docs/git/配置` 目录的文件树
- 支持文件预览（点击展示文件内容）
- 可选：支持刷新、导入、导出等操作

**组件**: 新建 `FileLibraryPanel.vue`

#### 页面 4：预览（preview）
**当前路径**: `/frontend/src/components/ContentSettingsPage.vue` 第 264-278 行

**改动**: 保持现有功能不变

---

### 3.2 前端数据结构调整

#### TOKEN_COLUMNS 更新
```javascript
const TOKEN_COLUMNS = [
  [
    { t: '{关键词}', d: '关键词库里的当前一条' },
    { t: '{图片}', d: '从图片库随机取一条，包成 Markdown 图片' },
    { t: '{时间1}', d: '17:42:30' },
    { t: '{AI扩写}', d: 'AI 扩写（占用）' },
  ],
  [
    { t: '{字符=5}', d: '5 位随机字母或数字' },
    { t: '{文章名}', d: '同一篇里与 {文章} 取自同一份素材的文件名' },
    { t: '{时间3}', d: '17时42分' },
    { t: '{文章}', d: '从文章库随机取一篇的正文' },
  ],
  [
    { t: '{数字=5}', d: '5 位随机数字' },
    { t: '{变量1}', d: '从变量库 1 随机抽一行' },
    { t: '{日期1}', d: '2026-08-25' },
    { t: '{变量3}', d: '从变量库 3 随机抽一行' },
  ],
]
```

#### 标签页配置
```javascript
const tabs = computed(() => [
  { key: 'template', label: '模板' },
  { key: 'bank', label: '变量设置' },
  { key: 'library', label: '文件库' },
  { key: 'preview', label: `预览 ${props.drafts.length}` },
])
```

---

### 3.3 文件库组件（FileLibraryPanel.vue）

**功能**:
- 读取 `/docs/git/配置/` 目录
- 树形展示文件结构
- 点击文件显示预览
- 支持文件操作（可选）

**状态管理**:
- selectedFile: 当前选中的文件路径
- files: 目录文件树结构
- fileContent: 预览内容

**布局**:
```
┌─ 文件树（左） ─┬─ 文件预览（右） ┐
│  📂 配置/     │ 内容显示区域    │
│    📄 xxx.txt │                │
│    📁 文章/   │                │
└───────────────┴────────────────┘
```

---

### 3.4 配置规则定义（/docs/git/tags.json）

**新增文件**: `/docs/git/tags.json`

**内容**: 定义所有支持的标签类型、对应 token、描述等

```json
{
  "version": "1.0",
  "tags": [
    {
      "id": "keyword",
      "category": "基础",
      "tokens": ["{关键词}"],
      "description": "从关键词库随机抽取",
      "configFile": "关键词.txt"
    },
    {
      "id": "random_char",
      "category": "随机字符",
      "tokens": ["{字符=N}", "{数字=N}", "{英文=N}", "{大写=N}", "{小写=N}", "{中文=N}"],
      "description": "N 位随机字符"
    },
    {
      "id": "media",
      "category": "媒体",
      "tokens": ["{图片}", "{文章}", "{文章名}"],
      "description": "图片库、文章库",
      "configFiles": ["文章/"]
    },
    {
      "id": "variables",
      "category": "变量库",
      "tokens": ["{变量1}", "{变量2}", "{变量3}", "{变量4}", "{变量5}"],
      "description": "从变量库随机抽取",
      "configDir": "变量/"
    },
    {
      "id": "datetime",
      "category": "时间日期",
      "tokens": ["{时间1}", "{时间2}", "{时间3}", "{时间4}", "{日期1}", "{日期2}", "{日期3}", "{日期4}"],
      "description": "时间和日期"
    },
    {
      "id": "ai_expand",
      "category": "AI",
      "tokens": ["{AI扩写}"],
      "description": "AI 扩写占用位置",
      "implemented": false
    }
  ]
}
```

---

## 4. 技术细节

### 4.1 样式调整

#### 选择器样式
- 添加 `::after` 伪元素显示「▼」箭头
- 或直接修改 select 元素的 `background-image` 属性

```css
.select {
  background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor'%3e%3cpath d='M6 9l6 6 6-6'/%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 8px center;
  background-size: 16px;
  padding-right: 28px;
}
```

#### 文件库按钮样式
```css
.btn-library {
  background: white;
  color: #333;
  border: 1px solid #ddd;
}

.btn-library:hover {
  border-color: #1677ff;
  color: #1677ff;
}
```

### 4.2 组件 Props 调整

```javascript
const props = defineProps({
  // 现有 props 保持不变
  titleTemplate: { type: String, required: true },
  bodyTemplates: { type: Array, required: true },
  keywordsText: { type: String, required: true },
  imagesText: { type: String, required: true },
  articleCount: { type: Number, required: true },
  varTexts: { type: Array, required: true },
  keywordOrder: { type: String, required: true },
  keywordTransform: { type: String, required: true },
  shuffleParagraphs: { type: Boolean, required: true },
  drafts: { type: Array, required: true },
  
  // 新增 props（如果文件库需要从父组件获取数据）
  configPath: { type: String, default: '/docs/git/配置' },
  tagConfig: { type: Object, default: () => ({}) },
})
```

### 4.3 Emit 事件

```javascript
const emit = defineEmits([
  // 现有事件保持不变
  'update:titleTemplate',
  'update:bodyTemplates',
  'update:keywordsText',
  'update:imagesText',
  'update:varTexts',
  'update:keywordOrder',
  'update:keywordTransform',
  'update:shuffleParagraphs',
  'open-dir',
  'import-text',
  'copy-draft',
  'copy-token',
  
  // 新增事件
  'library-action', // 文件库操作（刷新、导入等）
])
```

---

## 5. 现有代码保留

### 不改动的部分
- 「变量设置」标签页的所有功能和布局（第 197-262 行）
- 「预览」标签页的所有功能（第 264-278 行）
- 所有 CSS 变量和主题系统（--surface, --accent 等）
- 所有现有的 emit 事件和父组件交互逻辑

---

## 6. 实现顺序

### 第一阶段：UI 调整（当前 Vue 文件）
1. 更新 TOKEN_COLUMNS 数据结构，新增 {AI扩写}，调整排列
2. 修改 legend 字段标签文本（{标题}、【内容】）
3. 添加「文件库」按钮和样式
4. 添加「文件库」标签页到 tabs 配置
5. 在 template 中添加文件库标签页的条件渲染块

### 第二阶段：新组件和功能
6. 创建 FileLibraryPanel.vue 组件
7. 实现文件树展示逻辑
8. 集成到 ContentSettingsPage 的标签页中

### 第三阶段：配置和数据
9. 在 `/docs/git/` 创建 tags.json 文件
10. 整理现有配置文件的索引

---

## 7. 验收标准

- ✅ UI 字段标签已更新为参考设计中的文本
- ✅ {AI扩写} 标签出现在标签面板中
- ✅ 标签面板的三列布局按照参考设计排列
- ✅ 下拉选择器显示「▼」箭头
- ✅ 「文件库」按钮可见且样式正确
- ✅ 「文件库」标签页可以打开并显示目录结构
- ✅ `/docs/git/tags.json` 文件已创建
- ✅ 现有功能无回归（变量设置、预览等保持可用）

---

## 8. 风险与注意事项

- **文件库路径**: 需要确保 `/docs/git/配置/` 目录在所有环境中都存在
- **权限问题**: 如果需要编辑文件库中的文件，需要确保有写入权限
- **性能**: 大量文件时的树展示性能，可能需要虚拟滚动

---

## 9. 未来扩展

- 文件库的编辑功能
- {AI扩写} 的后端实现
- 标签配置的动态加载
- 文件库的搜索和过滤

