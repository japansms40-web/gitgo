<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { ListConfigTree, ReadConfigFile, ReadConfigFileTail, WriteConfigFile, RevealConfigFile } from '../../wailsjs/go/main/App'

const props = defineProps({
  openFile: { type: Object, default: null }, // 外部要求打开的文件：{ path, n }
})
const emit = defineEmits(['file-selected', 'open-dir'])

const tree = ref([])
const selectedFile = ref(null)
const baseline = ref('') // 磁盘上的原始内容，用来判断是否改动过
const editContent = ref('') // 文本框里正在编辑的内容
const truncated = ref(false) // 文件过大被截断：此时禁止编辑，以免保存丢内容
const tailView = ref(false) // 当前预览是否只加载了文件尾部（查看链接等大文件）
const loading = ref(false)
const saving = ref(false)
const errorMsg = ref('')
const expandedDirs = ref(new Set())

const dirty = computed(() => selectedFile.value !== null && editContent.value !== baseline.value)
const editable = computed(() => selectedFile.value !== null && !truncated.value)

// 把目录树按当前展开状态拍平成一串可见行，避免写递归组件；每行带 depth 缩进。
const visibleRows = computed(() => {
  const rows = []
  const walk = (nodes, depth) => {
    for (const node of nodes) {
      rows.push({ node, depth })
      if (node.isDir && expandedDirs.value.has(node.path)) {
        walk(node.children || [], depth + 1)
      }
    }
  }
  walk(tree.value, 0)
  return rows
})

onMounted(() => {
  loadFiles()
})

// 外部（如「查看链接」）要求打开某文件：刷新目录（新写入的文件才会出现），再选中并预览。
// immediate：文件库是按需挂载的，点「查看链接」时它才刚挂载，需在挂载即按当前 openFile 执行。
watch(
  () => props.openFile,
  async (v) => {
    if (!v || !v.path) return
    await loadFiles()
    const node = findNode(tree.value, v.path)
    if (node) await selectFile(node, !!v.tail)
    else errorMsg.value = `未找到文件：${v.path}`
  },
  { immediate: true },
)

// findNode 在目录树里按 path 找一个文件节点（深度优先，跳过目录）。
function findNode(nodes, path) {
  for (const node of nodes) {
    if (!node.isDir && node.path === path) return node
    if (node.isDir) {
      const hit = findNode(node.children || [], path)
      if (hit) return hit
    }
  }
  return null
}

async function loadFiles() {
  loading.value = true
  errorMsg.value = ''
  try {
    tree.value = (await ListConfigTree()) ?? []
    // 默认展开顶层目录，一进来就能看到变量/文章下的文件。
    for (const node of tree.value) {
      if (node.isDir) expandedDirs.value.add(node.path)
    }
  } catch (error) {
    errorMsg.value = String(error?.message ?? error)
    tree.value = []
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

async function onRowClick(node) {
  if (node.isDir) {
    toggleDir(node.path)
    return
  }
  await selectFile(node)
}

// tail：文件过大时只加载尾部最新内容（查看链接等大文件），否则加载头部。
async function selectFile(node, tail = false) {
  selectedFile.value = node.path
  baseline.value = ''
  editContent.value = '加载中…'
  truncated.value = false
  tailView.value = tail
  errorMsg.value = ''
  try {
    const preview = tail ? await ReadConfigFileTail(node.path) : await ReadConfigFile(node.path)
    baseline.value = preview.content ?? ''
    editContent.value = baseline.value
    truncated.value = !!preview.truncated
    emit('file-selected', node)
  } catch (error) {
    editContent.value = ''
    errorMsg.value = String(error?.message ?? error)
  }
}

// revealFile 在系统文件管理器里定位并选中当前文件，供大文件用外部工具打开（不在应用内加载）。
async function revealFile() {
  if (!selectedFile.value) return
  errorMsg.value = ''
  const err = await RevealConfigFile(selectedFile.value)
  if (err) errorMsg.value = err
}

async function saveFile() {
  if (!editable.value || !dirty.value || saving.value) return
  saving.value = true
  errorMsg.value = ''
  try {
    const err = await WriteConfigFile(selectedFile.value, editContent.value)
    if (err) {
      errorMsg.value = err
      return
    }
    baseline.value = editContent.value // 落盘成功，更新基线
  } catch (error) {
    errorMsg.value = String(error?.message ?? error)
  } finally {
    saving.value = false
  }
}

function refresh() {
  selectedFile.value = null
  baseline.value = ''
  editContent.value = ''
  truncated.value = false
  tailView.value = false
  loadFiles()
}
</script>

<template>
  <div class="file-library">
    <div class="library-toolbar">
      <button class="btn-action" @click="refresh">🔄 刷新</button>
      <button class="btn-action" @click="emit('open-dir')">📂 打开目录</button>
      <span class="toolbar-path mono">素材目录 · 点文件可直接编辑</span>
    </div>

    <div class="library-body">
      <div class="file-tree">
        <div class="tree-header">配置目录</div>
        <div v-if="loading" class="tree-tip">加载中…</div>
        <div v-else-if="errorMsg" class="tree-tip is-error">{{ errorMsg }}</div>
        <div v-else-if="visibleRows.length === 0" class="tree-tip">目录为空</div>
        <ul v-else class="tree-list">
          <li
            v-for="{ node, depth } in visibleRows"
            :key="node.path"
            class="tree-row"
            :class="{ 'is-selected': !node.isDir && selectedFile === node.path }"
            :style="{ paddingLeft: 8 + depth * 14 + 'px' }"
            @click="onRowClick(node)"
          >
            <span class="tree-caret">{{ node.isDir ? (expandedDirs.has(node.path) ? '▾' : '▸') : '' }}</span>
            <span class="tree-icon">{{ node.isDir ? '📁' : '📄' }}</span>
            <span class="tree-name">{{ node.name }}</span>
          </li>
        </ul>
      </div>

      <div class="file-preview">
        <div class="preview-header">
          <span v-if="selectedFile" class="file-path mono">{{ selectedFile }}</span>
          <span v-else class="placeholder">选择文件查看内容</span>
          <span v-if="dirty" class="preview-dirty">● 未保存</span>
          <span v-if="truncated" class="preview-truncated">{{ tailView ? '仅显示文件末尾最新部分 · 只读' : '文件过大，只读预览' }}</span>
          <div class="spacer" />
          <button v-if="selectedFile" class="btn-action btn-reveal" @click="revealFile">📂 在文件夹中显示</button>
          <button
            v-if="selectedFile"
            class="btn-save"
            :disabled="!editable || !dirty || saving"
            @click="saveFile"
          >
            {{ saving ? '保存中…' : '保存' }}
          </button>
        </div>
        <textarea
          v-if="selectedFile"
          class="preview-editor mono"
          :class="{ 'is-readonly': !editable }"
          :readonly="!editable"
          spellcheck="false"
          v-model="editContent"
        />
        <div v-else class="preview-empty">选择左侧文件查看并编辑内容</div>
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
  align-items: center;
  gap: 8px;
}
.toolbar-path {
  font-size: 11.5px;
  color: var(--muted);
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

.tree-tip {
  padding: 12px;
  font-size: 12px;
  color: var(--muted);
  text-align: center;
  word-break: break-all;
}
.tree-tip.is-error {
  color: var(--err, #d9534f);
}

.tree-list {
  margin: 0;
  padding: 0 0 8px;
  list-style: none;
}
.tree-row {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 26px;
  padding-right: 8px;
  font-size: 12.5px;
  color: var(--text);
  cursor: pointer;
  user-select: none;
  white-space: nowrap;
  overflow: hidden;
}
.tree-row:hover {
  background: var(--surface);
}
.tree-row.is-selected {
  background: var(--accent-weak);
  color: var(--accent);
}
.tree-caret {
  flex: 0 0 12px;
  font-size: 10px;
  color: var(--muted);
  text-align: center;
}
.tree-icon {
  flex: 0 0 auto;
  font-size: 12px;
}
.tree-name {
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-preview {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.preview-header {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  color: var(--text);
}
.preview-truncated {
  font-size: 11px;
  color: var(--warn, #9a6700);
}
.preview-dirty {
  font-size: 11px;
  color: var(--accent);
}
.preview-header .spacer {
  flex: 1;
}
.btn-save {
  padding: 5px 16px;
  border: none;
  border-radius: 4px;
  background: var(--accent);
  color: #fff;
  font-size: 12px;
  cursor: pointer;
}
.btn-save:disabled {
  opacity: 0.45;
  cursor: default;
}

.placeholder {
  color: var(--muted);
}

.file-path {
  font-family: monospace;
  color: var(--accent);
}

.preview-editor {
  flex: 1;
  margin: 0;
  padding: 16px;
  border: none;
  outline: none;
  resize: none;
  overflow: auto;
  font-size: 12px;
  line-height: 1.6;
  background: var(--surface);
  color: var(--text);
  white-space: pre;
  tab-size: 4;
}
.preview-editor.is-readonly {
  background: var(--surface-2);
  color: var(--muted);
}

.preview-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--muted);
}
</style>
