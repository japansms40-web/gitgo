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
