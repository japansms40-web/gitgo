<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import TitleBar from './components/TitleBar.vue'
import NavRail from './components/NavRail.vue'
import PublishPage from './components/PublishPage.vue'
import HelpPage from './components/HelpPage.vue'
import RunParamsPanel from './components/RunParamsPanel.vue'
import LogPanel from './components/LogPanel.vue'
import StatusBar from './components/StatusBar.vue'
import * as App from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

const APP_VERSION = 'v1.0.0'

const THEME_KEY = 'ghpublisher.theme'
const theme = ref(localStorage.getItem(THEME_KEY) || 'light')
function applyTheme() {
  document.documentElement.dataset.theme = theme.value
}
function toggleTheme() {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  localStorage.setItem(THEME_KEY, theme.value)
  applyTheme()
}
applyTheme()

const page = ref('publish')

const cfg = reactive({
  token: '',
  owner: '',
  repo: '',
  branch: 'main',
  dir: 'posts',
  autoCreate: false,
  intervalSec: 1,
  retries: 2,
})
const tokenStatus = ref('')
const queue = ref([]) // {title, repoPath, status, url}
const running = ref(false)
const logs = ref([]) // {time, tag, kind, msg, highlight}
const autoScroll = ref(true)
const elapsed = ref('00:00')
const banner = ref('')

let startedAt = 0
let timer = null

function formatElapsed(ms) {
  const total = Math.round(ms / 1000)
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  const pad = (n) => String(n).padStart(2, '0')
  return h > 0 ? `${pad(h)}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`
}

function showBanner(msg) {
  banner.value = msg
  setTimeout(() => {
    if (banner.value === msg) banner.value = ''
  }, 4000)
}

function pushLog(kind, tag, msg, highlight = false) {
  logs.value.push({ time: new Date().toTimeString().slice(0, 8), tag, kind, msg, highlight })
  if (logs.value.length > 1000) logs.value.shift()
}

const doneCount = computed(() => queue.value.filter((q) => q.status === '成功' || q.status.startsWith('失败')).length)
const failedCount = computed(() => queue.value.filter((q) => q.status.startsWith('失败')).length)
const successCount = computed(() => queue.value.filter((q) => q.status === '成功').length)

const pill = computed(() => {
  const total = queue.value.length
  const done = doneCount.value
  const failed = failedCount.value
  if (running.value) return { text: `发布中 · ${done}/${total}`, kind: 'running' }
  if (total === 0 || done === 0) return { text: '待机', kind: 'muted' }
  if (failed > 0) return { text: `已完成 · ${done - failed} 成功 · ${failed} 失败`, kind: 'error' }
  return { text: `已完成 · ${done}/${total}`, kind: 'success' }
})

onMounted(async () => {
  const loaded = await App.LoadConfig()
  Object.assign(cfg, loaded)

  EventsOn('publish:log', (line) => {
    logs.value.push(line)
    if (logs.value.length > 1000) logs.value.shift()
  })
  EventsOn('publish:status', (u) => {
    const item = queue.value[u.index]
    if (!item) return
    switch (u.kind) {
      case 'start':
        item.status = '发布中'
        break
      case 'success':
        item.status = '成功'
        item.url = u.url
        break
      case 'failure':
        item.status = '失败: ' + u.err
        break
      case 'retry':
        item.status = '重试中: ' + u.err
        break
    }
  })
  EventsOn('publish:done', (errMsg) => {
    running.value = false
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    elapsed.value = formatElapsed(Date.now() - startedAt)
    if (errMsg) showBanner(errMsg)
  })
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})

async function rescanQueue(paths) {
  if (!paths || paths.length === 0) return
  try {
    const items = await App.ScanQueue(paths, cfg.dir)
    queue.value = items.map((it) => ({ title: it.title, repoPath: it.repoPath, status: '待发布', url: '' }))
  } catch (e) {
    showBanner(String(e))
  }
}

async function onSelectFolder() {
  const dirPath = await App.SelectFolder()
  if (dirPath) await rescanQueue([dirPath])
}
async function onSelectFiles() {
  const files = await App.SelectFiles()
  if (files && files.length) await rescanQueue(files)
}
async function onClearQueue() {
  await App.ClearQueue()
  queue.value = []
}

async function onValidateToken() {
  if (!cfg.token) {
    showBanner('请先填写 Token')
    return
  }
  tokenStatus.value = '验证中…'
  try {
    const login = await App.ValidateToken(cfg.token)
    tokenStatus.value = '✓ 已登录: ' + login
  } catch (e) {
    tokenStatus.value = '无效: ' + e
  }
}

async function onStart() {
  if (queue.value.length === 0) {
    showBanner('队列为空，请先添加文件')
    return
  }
  const err = await App.StartPublish({ ...cfg })
  if (err) {
    showBanner(err)
    return
  }
  running.value = true
  startedAt = Date.now()
  elapsed.value = '00:00'
  pushLog('info', '[信息]', `开始发布 ${queue.value.length} 篇到 ${cfg.owner}/${cfg.repo}`)
  timer = setInterval(() => {
    elapsed.value = formatElapsed(Date.now() - startedAt)
  }, 1000)
}
async function onStop() {
  await App.StopPublish()
}

async function onSaveConfig() {
  const err = await App.SaveConfig({ ...cfg })
  if (err) {
    showBanner(err)
    return
  }
  pushLog('info', '[信息]', '配置已保存')
}

async function onExportLinks() {
  const urls = queue.value.filter((q) => q.status === '成功' && q.url).map((q) => q.url)
  if (urls.length === 0) {
    showBanner('暂无成功链接')
    return
  }
  const err = await App.ExportLinks(urls)
  if (err) showBanner(err)
}

function logText() {
  return logs.value.map((l) => `${l.time} ${l.tag} ${l.msg}`).join('\n')
}
async function onCopyLog() {
  await App.CopyToClipboard(logText())
}
async function onExportLog() {
  const err = await App.ExportLog(logText())
  if (err) showBanner(err)
}
function onClearLog() {
  logs.value = []
}
</script>

<template>
  <div class="shell">
    <TitleBar :theme="theme" :version="APP_VERSION" :pill-text="pill.text" :pill-kind="pill.kind" @toggle-theme="toggleTheme" />

    <div v-if="banner" class="banner">{{ banner }}</div>

    <div class="body">
      <NavRail :page="page" @navigate="(p) => (page = p)" />

      <PublishPage
        v-if="page === 'publish'"
        v-model:token="cfg.token"
        v-model:owner="cfg.owner"
        v-model:repo="cfg.repo"
        v-model:branch="cfg.branch"
        v-model:dir="cfg.dir"
        :token-status="tokenStatus"
        :queue="queue"
        @validate-token="onValidateToken"
        @select-folder="onSelectFolder"
        @select-files="onSelectFiles"
        @clear-queue="onClearQueue"
      />
      <HelpPage v-else />

      <RunParamsPanel
        v-model:interval-sec="cfg.intervalSec"
        v-model:retries="cfg.retries"
        v-model:auto-create="cfg.autoCreate"
        :queue-count="queue.length"
        :running="running"
        @start="onStart"
        @stop="onStop"
        @save-config="onSaveConfig"
        @clear-queue="onClearQueue"
        @export-links="onExportLinks"
      />
    </div>

    <LogPanel
      :lines="logs"
      v-model:auto-scroll="autoScroll"
      @copy="onCopyLog"
      @export="onExportLog"
      @clear="onClearLog"
    />
    <StatusBar
      :total="queue.length"
      :success="successCount"
      :fail="failedCount"
      :pending="queue.length - doneCount"
      :elapsed="elapsed"
    />
  </div>
</template>

<style scoped>
.shell {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg);
}
.banner {
  position: absolute;
  top: 60px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--err);
  color: #fff;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 12.5px;
  box-shadow: var(--shadow);
  z-index: 50;
}
.body {
  flex: 1;
  display: flex;
  min-height: 0;
}
</style>
