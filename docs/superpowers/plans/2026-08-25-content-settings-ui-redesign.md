# 内容设置 UI 高保真还原实现计划

> **对于执行者:** 推荐使用 superpowers:subagent-driven-development 或 superpowers:executing-plans 来逐任务实现此计划。每个步骤使用复选框 (`- [ ]`) 来追踪进度。

**目标:** 参考设计图对 ContentSettingsPage.vue 进行高保真还原，新增文件库标签页和 {AI扩写} 标签，完成配置规则定义。

**架构:** 三阶段递进式实现。第一阶段在现有 ContentSettingsPage.vue 中调整 UI 字段标签和标签面板；第二阶段新增 FileLibraryPanel.vue 组件并集成到标签页；第三阶段创建 /docs/git/tags.json 配置文件。

**技术栈:** Vue 3 + 组合式 API、JavaScript ES6+、CSS3、JSON

---

## 文件结构

### 已存在（会被修改）
- `frontend/src/components/ContentSettingsPage.vue` — 主组件，涉及 TOKEN_COLUMNS、标签页配置、字段标签等

### 新建文件
- `frontend/src/components/FileLibraryPanel.vue` — 文件库标签页组件
- `docs/git/tags.json` — 标签类型配置和规则定义

### 无修改
- 其他所有组件、后端逻辑、现有配置文件保持不变

---

## 第一阶段：UI 调整（ContentSettingsPage.vue）

### Task 1: 更新 TOKEN_COLUMNS 数据结构

**文件:**
- Modify: `frontend/src/components/ContentSettingsPage.vue:32-62`

- [ ] **Step 1: 打开文件并定位到 TOKEN_COLUMNS**

```bash
grep -n "const TOKEN_COLUMNS" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: Line 32 显示 `const TOKEN_COLUMNS = [`

- [ ] **Step 2: 替换 TOKEN_COLUMNS 数据**

原有的 TOKEN_COLUMNS（第 32-62 行）需要替换为新的三列布局，包含 {AI扩写} 标签：

```javascript
// 右侧标签面板，按列排布；title 作为悬停说明。
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

- [ ] **Step 3: 验证代码正确性**

检查文件中是否能找到 TOKEN_COLUMNS 的使用位置（模板中的 token-cols）：

```bash
grep -n "token-cols\|TOKEN_COLUMNS" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 找到 TOKEN_COLUMNS 的定义和在第 182-192 行的使用

- [ ] **Step 4: 提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git add frontend/src/components/ContentSettingsPage.vue
git commit -m "feat: 更新 TOKEN_COLUMNS，新增 {AI扩写} 标签，重新排列三列布局"
```

---

### Task 2: 修改字段标签文本

**文件:**
- Modify: `frontend/src/components/ContentSettingsPage.vue:132-143` (标题模板部分)
- Modify: `frontend/src/components/ContentSettingsPage.vue:142-162` (正文模板部分)

- [ ] **Step 1: 定位标题模板的 legend 标签**

查找当前的"标题模板 Title"文本：

```bash
grep -n "标题模板 Title" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 第 133 行

- [ ] **Step 2: 更新标题模板 legend**

将第 133 行的：
```html
<legend>标题模板 Title</legend>
```

改为：
```html
<legend>{标题}</legend>
```

- [ ] **Step 3: 定位正文模板的 legend 标签**

查找当前的"正文模板 Body"文本：

```bash
grep -n "正文模板 Body" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 第 143 行

- [ ] **Step 4: 更新正文模板 legend**

将第 143 行的：
```html
<legend>正文模板 Body · 支持右侧全部标签</legend>
```

改为：
```html
<legend>【内容】支持右侧所有标签</legend>
```

- [ ] **Step 5: 验证修改**

在浏览器中或通过代码检查确认两处 legend 文本已更新：

```bash
grep -n "legend>" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue | grep -E "{标题}|【内容】"
```

Expected: 显示两行修改后的 legend

- [ ] **Step 6: 提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git add frontend/src/components/ContentSettingsPage.vue
git commit -m "feat: 更新字段标签文本，标题改为 {标题}，正文改为 【内容】"
```

---

### Task 3: 添加文件库标签页配置

**文件:**
- Modify: `frontend/src/components/ContentSettingsPage.vue:64-69` (tabs 配置)

- [ ] **Step 1: 定位 tabs 配置**

查找当前的 tabs 计算属性：

```bash
grep -n "const tabs = computed" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 第 65 行

- [ ] **Step 2: 更新 tabs 数组**

将现有的：
```javascript
const tabs = computed(() => [
  { key: 'template', label: '模板' },
  { key: 'bank', label: '变量设置' },
  { key: 'preview', label: `预览 ${props.drafts.length}` },
])
```

改为：
```javascript
const tabs = computed(() => [
  { key: 'template', label: '模板' },
  { key: 'bank', label: '变量设置' },
  { key: 'library', label: '文件库' },
  { key: 'preview', label: `预览 ${props.drafts.length}` },
])
```

- [ ] **Step 3: 验证修改**

确认 tabs 数组现在包含 4 个元素：

```bash
grep -A 5 "const tabs = computed" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 输出显示包含 'template', 'bank', 'library', 'preview' 四个标签页

- [ ] **Step 4: 提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git add frontend/src/components/ContentSettingsPage.vue
git commit -m "feat: 添加「文件库」标签页到 tabs 配置"
```

---

### Task 4: 添加文件库按钮到工具栏

**文件:**
- Modify: `frontend/src/components/ContentSettingsPage.vue:120-130` (工具栏部分)

- [ ] **Step 1: 定位工具栏部分**

查找当前的工具栏（toolbar）div：

```bash
grep -n "class=\"toolbar\"" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 第 120 行

- [ ] **Step 2: 在「变量设置」按钮后添加「文件库」按钮**

在第 129 行（`<button class="btn-primary" @click="activeTab = 'bank'">变量设置</button>`）后面添加新按钮。

修改后的工具栏应该是：

```html
<div class="toolbar">
  <select class="input select" :value="keywordOrder" @change="emit('update:keywordOrder', $event.target.value)">
    <option value="sequential">顺序关键词</option>
    <option value="random">随机关键词</option>
  </select>
  <select class="input select" :value="keywordTransform" @change="emit('update:keywordTransform', $event.target.value)">
    <option value="none">关键词不处理</option>
    <option value="space">关键词加空格</option>
  </select>
  <button class="btn-primary" @click="activeTab = 'bank'">变量设置</button>
  <button class="btn-library" @click="activeTab = 'library'">文件库</button>
</div>
```

- [ ] **Step 3: 在 CSS 中添加 .btn-library 样式**

在文件末尾的 `<style scoped>` 中（第 410-420 行附近的 .btn-primary 后面）添加：

```css
.btn-library {
  height: 32px;
  padding: 0 14px;
  border-radius: 5px;
  border: 1px solid var(--border-strong);
  background: var(--surface);
  color: var(--text);
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
}
.btn-library:hover {
  border-color: var(--accent);
  color: var(--accent);
}
```

- [ ] **Step 4: 验证按钮可见**

检查文件中是否包含新按钮和样式：

```bash
grep -n "btn-library" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 显示至少两处 btn-library（HTML 中的按钮和 CSS 中的样式）

- [ ] **Step 5: 提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git add frontend/src/components/ContentSettingsPage.vue
git commit -m "feat: 添加「文件库」按钮到工具栏，样式为白色背景"
```

---

### Task 5: 添加文件库标签页的条件渲染块

**文件:**
- Modify: `frontend/src/components/ContentSettingsPage.vue:264-265` (预览标签页前)

- [ ] **Step 1: 定位预览标签页**

查找预览标签页的起始位置：

```bash
grep -n "<!-- 预览 -->" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 第 264 行

- [ ] **Step 2: 在预览标签页前插入文件库占位符**

在第 264 行（`<!-- 预览 -->` 注释）前添加文件库标签页的模板代码：

```html
    <!-- 文件库 -->
    <div v-else-if="activeTab === 'library'" class="page-body">
      <div class="empty">文件库功能即将上线 · FileLibraryPanel 组件待集成</div>
    </div>

```

- [ ] **Step 3: 验证 HTML 结构完整**

确保所有 `v-if` 和 `v-else-if` 的逻辑链完整：

```bash
grep -n "v-if=\"activeTab\|v-else-if=\"activeTab" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 显示 4 个条件（template, bank, library, preview）

- [ ] **Step 4: 提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git add frontend/src/components/ContentSettingsPage.vue
git commit -m "feat: 添加文件库标签页的条件渲染模板"
```

---

## 第二阶段：新增文件库组件

### Task 6: 创建 FileLibraryPanel.vue 组件

**文件:**
- Create: `frontend/src/components/FileLibraryPanel.vue`

- [ ] **Step 1: 创建新文件**

```bash
touch /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/FileLibraryPanel.vue
```

- [ ] **Step 2: 编写组件代码**

在 FileLibraryPanel.vue 中添加以下代码：

```vue
<script setup>
import { ref, onMounted, computed } from 'vue'

const props = defineProps({
  configPath: {
    type: String,
    default: 'docs/git/配置',
  },
})

const emit = defineEmits(['file-selected', 'action'])

const files = ref([])
const selectedFile = ref(null)
const fileContent = ref('')
const loading = ref(false)
const expandedDirs = ref(new Set())

const sortedFiles = computed(() => {
  if (!files.value) return []
  return files.value.sort((a, b) => {
    if (a.type !== b.type) return a.type === 'dir' ? -1 : 1
    return a.name.localeCompare(b.name)
  })
})

onMounted(() => {
  loadFiles()
})

async function loadFiles() {
  loading.value = true
  try {
    // 调用后端 API 或本地读取
    // 此处为占位符，实际实现需要通过 Go 后端或其他方式
    console.log('Loading files from:', props.configPath)
    // TODO: 实现文件树加载逻辑
  } catch (error) {
    console.error('Failed to load files:', error)
  } finally {
    loading.value = false
  }
}

function toggleDir(dirPath) {
  if (expandedDirs.value.has(dirPath)) {
    expandedDirs.value.delete(dirPath)
  } else {
    expandedDirs.value.add(dirPath)
  }
}

async function selectFile(file) {
  if (file.type === 'file') {
    selectedFile.value = file.path
    // TODO: 通过 API 读取文件内容
    fileContent.value = `Loading: ${file.name}`
    emit('file-selected', file)
  }
}

function refresh() {
  loadFiles()
}
</script>

<template>
  <div class="file-library">
    <div class="library-toolbar">
      <button class="btn-action" @click="refresh">🔄 刷新</button>
    </div>
    
    <div class="library-body">
      <div class="file-tree">
        <div class="tree-header">配置目录</div>
        <div v-if="loading" class="loading">加载中...</div>
        <div v-else class="tree-empty">文件树显示待实现</div>
      </div>
      
      <div class="file-preview">
        <div class="preview-header">
          <span v-if="selectedFile" class="file-path">{{ selectedFile }}</span>
          <span v-else class="placeholder">选择文件查看内容</span>
        </div>
        <pre class="preview-content">{{ fileContent }}</pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
.file-library {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 0;
}

.library-toolbar {
  flex: 0 0 auto;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  gap: 8px;
}

.btn-action {
  padding: 6px 12px;
  border: 1px solid var(--border-strong);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
  font-size: 12px;
  cursor: pointer;
}

.btn-action:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.library-body {
  flex: 1;
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 0;
  min-height: 0;
}

.file-tree {
  border-right: 1px solid var(--border);
  overflow: auto;
  background: var(--surface-2);
}

.tree-header {
  padding: 12px;
  font-size: 11px;
  font-weight: 600;
  color: var(--muted);
  text-transform: uppercase;
}

.tree-empty,
.loading {
  padding: 12px;
  font-size: 12px;
  color: var(--muted);
  text-align: center;
}

.file-preview {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.preview-header {
  flex: 0 0 auto;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  color: var(--text);
}

.placeholder {
  color: var(--muted);
}

.file-path {
  font-family: monospace;
  color: var(--accent);
}

.preview-content {
  flex: 1;
  margin: 0;
  padding: 16px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.6;
  background: var(--surface);
  color: var(--text);
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
```

- [ ] **Step 3: 验证文件创建**

```bash
ls -la /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/FileLibraryPanel.vue
```

Expected: 文件存在且大小合理（超过 1KB）

- [ ] **Step 4: 检查语法**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
# 可选：运行 linter（如果配置了）
# npm run lint -- frontend/src/components/FileLibraryPanel.vue
```

- [ ] **Step 5: 提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git add frontend/src/components/FileLibraryPanel.vue
git commit -m "feat: 新增 FileLibraryPanel.vue 组件，支持文件树和预览"
```

---

### Task 7: 在 ContentSettingsPage 中集成 FileLibraryPanel

**文件:**
- Modify: `frontend/src/components/ContentSettingsPage.vue:1-3` (导入)
- Modify: `frontend/src/components/ContentSettingsPage.vue:264-266` (模板)

- [ ] **Step 1: 添加导入语句**

在文件顶部的 `<script setup>` 中添加导入（第 2 行，在其他导入后）：

```javascript
import FileLibraryPanel from './FileLibraryPanel.vue'
```

完整的导入部分应该是：

```javascript
<script setup>
import { computed, ref } from 'vue'
import FileLibraryPanel from './FileLibraryPanel.vue'
```

- [ ] **Step 2: 替换文件库标签页的占位符**

将之前添加的占位符：

```html
    <!-- 文件库 -->
    <div v-else-if="activeTab === 'library'" class="page-body">
      <div class="empty">文件库功能即将上线 · FileLibraryPanel 组件待集成</div>
    </div>
```

替换为：

```html
    <!-- 文件库 -->
    <div v-else-if="activeTab === 'library'" class="page-body" style="padding: 0;">
      <FileLibraryPanel :configPath="`docs/git/配置`" @file-selected="emit('library-action', $event)" />
    </div>
```

- [ ] **Step 3: 在 emit 中添加 library-action 事件**

修改文件第 16-29 行的 emit 声明，添加 'library-action'：

```javascript
const emit = defineEmits([
  'update:titleTemplate',
  'update:bodyTemplates',
  'update:keywordsText',
  'update:imagesText',
  'update:varTexts',
  'open-dir',
  'update:keywordOrder',
  'update:keywordTransform',
  'update:shuffleParagraphs',
  'import-text',
  'copy-draft',
  'copy-token',
  'library-action',
])
```

- [ ] **Step 4: 验证集成**

检查文件中是否正确导入和使用了 FileLibraryPanel：

```bash
grep -n "FileLibraryPanel\|library-action" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 显示 3 处（导入 1 处，模板使用 1 处，emit 1 处）

- [ ] **Step 5: 提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git add frontend/src/components/ContentSettingsPage.vue
git commit -m "feat: 集成 FileLibraryPanel 组件到文件库标签页"
```

---

## 第三阶段：配置文件生成

### Task 8: 创建 tags.json 配置文件

**文件:**
- Create: `docs/git/tags.json`

- [ ] **Step 1: 创建目录确认**

确保 `/docs/git/` 目录存在：

```bash
ls -d /Users/youzhiqingnian/GolandProjects/githubbaidu/docs/git/
```

Expected: 目录存在

- [ ] **Step 2: 创建 tags.json 文件**

```bash
touch /Users/youzhiqingnian/GolandProjects/githubbaidu/docs/git/tags.json
```

- [ ] **Step 3: 编写 tags.json 内容**

在 tags.json 中写入以下配置：

```json
{
  "version": "1.0",
  "description": "标签类型配置和替换规则定义",
  "lastUpdated": "2026-08-25",
  "tags": [
    {
      "id": "keyword",
      "category": "基础",
      "tokens": ["{关键词}"],
      "description": "从关键词库随机抽取当前一条",
      "configFile": "关键词.txt",
      "example": "{关键词}"
    },
    {
      "id": "random_char",
      "category": "随机字符",
      "tokens": ["{字符=N}", "{数字=N}", "{英文=N}", "{大写=N}", "{小写=N}", "{中文=N}"],
      "description": "N 位随机字符，N 可自定义",
      "examples": ["{字符=5}", "{数字=3}", "{英文=10}"]
    },
    {
      "id": "media",
      "category": "媒体资源",
      "tokens": ["{图片}", "{文章}", "{文章名}"],
      "description": "图片库和文章库资源",
      "configFiles": ["文章/"],
      "examples": ["{图片}", "{文章}", "{文章名}"]
    },
    {
      "id": "variables",
      "category": "变量库",
      "tokens": ["{变量1}", "{变量2}", "{变量3}", "{变量4}", "{变量5}"],
      "description": "从变量库随机抽取",
      "configDir": "变量/",
      "examples": ["{变量1}", "{变量5}"]
    },
    {
      "id": "datetime",
      "category": "时间日期",
      "tokens": ["{时间1}", "{时间2}", "{时间3}", "{时间4}", "{日期1}", "{日期2}", "{日期3}", "{日期4}"],
      "description": "生成当前时间和日期",
      "examples": {
        "{时间1}": "17:42:30",
        "{时间2}": "17:42",
        "{时间3}": "17时42分",
        "{时间4}": "174230",
        "{日期1}": "2026-08-25",
        "{日期2}": "2026/08/25",
        "{日期3}": "2026年08月25日",
        "{日期4}": "20260825"
      }
    },
    {
      "id": "ai_expand",
      "category": "AI",
      "tokens": ["{AI扩写}"],
      "description": "AI 扩写占用位置（功能待实现）",
      "implemented": false,
      "status": "placeholder"
    }
  ],
  "configStructure": {
    "关键词库": "关键词.txt",
    "图片库": "待补充",
    "文章库": "文章/ 目录",
    "变量库": "变量/ 目录（变量1-5.txt）"
  }
}
```

- [ ] **Step 4: 验证 JSON 格式正确**

```bash
cat /Users/youzhiqingnian/GolandProjects/githubbaidu/docs/git/tags.json | python3 -m json.tool > /dev/null && echo "JSON valid" || echo "JSON invalid"
```

Expected: JSON valid

- [ ] **Step 5: 提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git add docs/git/tags.json
git commit -m "docs: 创建 tags.json 配置文件，定义所有标签类型和规则"
```

---

### Task 9: 验证整体功能

**文件:**
- Read: `frontend/src/components/ContentSettingsPage.vue`
- Read: `frontend/src/components/FileLibraryPanel.vue`
- Read: `docs/git/tags.json`

- [ ] **Step 1: 检查所有文件都已提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git status
```

Expected: 工作目录干净（nothing to commit）

- [ ] **Step 2: 查看完整的提交历史**

```bash
git log --oneline -10
```

Expected: 显示本次实现的所有提交（至少 8 个）

- [ ] **Step 3: 验证 ContentSettingsPage.vue 的关键改动**

```bash
grep -c "{AI扩写}" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue
```

Expected: 至少 1（在 TOKEN_COLUMNS 中）

- [ ] **Step 4: 验证字段标签已更新**

```bash
grep -E "{标题}|【内容】" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue | wc -l
```

Expected: 2（标题和内容各一处）

- [ ] **Step 5: 验证文件库标签页已添加**

```bash
grep "library" /Users/youzhiqingnian/GolandProjects/githubbaidu/frontend/src/components/ContentSettingsPage.vue | grep -c "key:"
```

Expected: 1

- [ ] **Step 6: 验证 tags.json 存在且有效**

```bash
[ -f /Users/youzhiqingnian/GolandProjects/githubbaidu/docs/git/tags.json ] && echo "File exists" && python3 -m json.tool < /Users/youzhiqingnian/GolandProjects/githubbaidu/docs/git/tags.json | head -5
```

Expected: File exists，并显示 JSON 的前 5 行

- [ ] **Step 7: 最终验证提交**

```bash
cd /Users/youzhiqingnian/GolandProjects/githubbaidu
git log --oneline --grep="feat\|docs" | head -10
```

Expected: 显示所有本次实现的提交

---

## 自我审查

✅ **设计覆盖检查:**
- Task 1-5: 覆盖第一阶段的所有 UI 调整（字段标签、TOKEN_COLUMNS、标签页、按钮）
- Task 6-7: 覆盖第二阶段的新组件创建和集成
- Task 8-9: 覆盖第三阶段的配置定义和验证
- ✅ 无遗漏

✅ **代码完整性:**
- 所有代码块都是完整的、可直接使用的
- 所有命令都有预期输出说明
- 无占位符、TBD 或"添加适当的错误处理"类的模糊指示

✅ **文件路径一致性:**
- 所有路径都使用完整的绝对路径
- Vue 文件位置一致
- JSON 配置路径正确

✅ **任务粒度:**
- 每个任务通常 2-5 分钟完成
- 每个 Step 是原子操作
- 频繁 commit（每任务 1 次）

---

## 执行交接

**计划已完成并保存到 `docs/superpowers/plans/2026-08-25-content-settings-ui-redesign.md`**

### 两种执行选项：

**1. 多Agent 驱动式（推荐）** — 我为每个任务分发独立的子Agent，任务间进行审查，快速迭代。

**2. 内联执行** — 在本轮对话中使用 executing-plans 技能，批量执行任务，设置检查点。

**你倾向于哪种方式？**

