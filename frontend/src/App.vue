<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import TitleBar from './components/TitleBar.vue'
import NavRail from './components/NavRail.vue'
import PublishPage from './components/PublishPage.vue'
import HelpPage from './components/HelpPage.vue'
import RunParamsPanel from './components/RunParamsPanel.vue'
import LogPanel from './components/LogPanel.vue'
import StatusBar from './components/StatusBar.vue'
import ResultsModal from './components/ResultsModal.vue'
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
  threads: 1,
  intervalSec: 1,
  perAccountCount: 1,
  failSwitchCount: 3,
  cycleRounds: 1,
  roundIntervalSec: 1,
  keywordSlots: 0,
  createRepo: false,
})
const accounts = ref([]) // {ck, ua, ip, status, success, fail, total, bad}
const contentCount = ref(0)
const running = ref(false)
const paused = ref(false)
const round = ref(0)
const roundDone = ref(0)
const roundTotal = ref(0)
const logs = ref([])
const autoScroll = ref(true)
const elapsed = ref('00:00')
const banner = ref('')
const showResults = ref(false)
const results = ref([])

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

const successCount = computed(() => accounts.value.filter((a) => a.status === '成功').length)
const failedCount = computed(() => accounts.value.filter((a) => a.status === '失败').length)
const pendingCount = computed(() => accounts.value.length - successCount.value - failedCount.value)

const pill = computed(() => {
  const total = accounts.value.length
  const done = successCount.value + failedCount.value
  if (running.value) return { text: `发布中 · ${done}/${total}`, kind: 'running' }
  if (total === 0 || done === 0) return { text: '待机', kind: 'muted' }
  if (failedCount.value > 0) return { text: `已完成 · ${successCount.value} 成功 · ${failedCount.value} 失败`, kind: 'error' }
  return { text: `已完成 · ${done}/${total}`, kind: 'success' }
})

onMounted(async () => {
  const loaded = await App.LoadConfig()
  Object.assign(cfg, loaded)
  accounts.value = await App.LoadAccounts()

  EventsOn('publish:log', (line) => {
    logs.value.push(line)
    if (logs.value.length > 1000) logs.value.shift()
  })
  EventsOn('publish:account', (u) => {
    const item = accounts.value[u.index]
    if (!item) return
    item.status = u.status
    item.success = u.success
    item.fail = u.fail
    item.total = u.total
  })
  EventsOn('publish:round', (u) => {
    round.value = u.round
    roundDone.value = u.done
    roundTotal.value = u.total
  })
  EventsOn('publish:done', (errMsg) => {
    running.value = false
    paused.value = false
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

// ---- 发布内容（本地 Markdown/文本） ----
async function rescanContent(paths) {
  if (!paths || paths.length === 0) return
  try {
    const items = await App.ScanQueue(paths)
    contentCount.value = items.length
  } catch (e) {
    showBanner(String(e))
  }
}
async function onSelectContentFolder() {
  const dirPath = await App.SelectFolder()
  if (dirPath) await rescanContent([dirPath])
}
async function onSelectContentFiles() {
  const files = await App.SelectFiles()
  if (files && files.length) await rescanContent(files)
}
async function onClearContent() {
  await App.ClearQueue()
  contentCount.value = 0
}

// ---- 账号队列 ----
async function onImportAccountClick() {
  try {
    const paths = await App.SelectAccountFiles()
    if (paths && paths.length) accounts.value = await App.ImportAccountsFile(paths)
  } catch (e) {
    showBanner(String(e))
  }
}
async function onImportAccountFiles(paths) {
  try {
    accounts.value = await App.ImportAccountsFile(paths)
  } catch (e) {
    showBanner(String(e))
  }
}
async function onPasteClipboard() {
  try {
    accounts.value = await App.PasteAccountsFromClipboard()
  } catch (e) {
    showBanner(String(e))
  }
}
async function onRemoveAccount(index) {
  try {
    accounts.value = await App.RemoveAccount(index)
  } catch (e) {
    showBanner(String(e))
  }
}
async function onMarkBad(index) {
  try {
    accounts.value = await App.MarkBadAccount(index)
  } catch (e) {
    showBanner(String(e))
  }
}
async function onClearAccounts() {
  accounts.value = await App.ClearAccounts()
}
async function onExportResult() {
  const err = await App.ExportAccountsResult()
  if (err) showBanner(err)
}
async function onTestAccount(index) {
  try {
    const updated = await App.TestAccount(index)
    accounts.value[index] = updated
  } catch (e) {
    showBanner(String(e))
  }
}
async function onCopyCK(ck) {
  if (!ck) return
  await App.CopyToClipboard(ck)
}

async function onStart() {
  if (contentCount.value === 0) {
    showBanner('请先选择要发布的内容')
    return
  }
  if (accounts.value.filter((a) => !a.bad).length === 0) {
    showBanner('账号队列为空（或都已标记为坏号），请先导入账号')
    return
  }
  const err = await App.StartPublish({ ...cfg })
  if (err) {
    showBanner(err)
    return
  }
  running.value = true
  paused.value = false
  round.value = 0
  roundDone.value = 0
  roundTotal.value = 0
  startedAt = Date.now()
  elapsed.value = '00:00'
  pushLog('info', '[信息]', `开始发布，账号 ${accounts.value.length} 个，内容 ${contentCount.value} 篇`)
  timer = setInterval(() => {
    elapsed.value = formatElapsed(Date.now() - startedAt)
  }, 1000)
}
async function onPause() {
  await App.PausePublish()
  paused.value = true
}
async function onResume() {
  await App.ResumePublish()
  paused.value = false
}
async function onStop() {
  await App.StopPublish()
}

function onKeywordSettings() {
  showBanner('关键词设置功能待实现，需要你进一步说明关键词库和插入规则')
}
function onSwitchProfile() {
  showBanner('换号特征编辑功能待实现，需要你进一步说明具体字段')
}
async function onViewResults() {
  results.value = await App.GetPublishResults()
  showResults.value = true
}
async function onCopyAllResults() {
  const text = results.value.map((r) => `${r.time} ${r.ck} ${r.title} ${r.value}`).join('\n')
  await App.CopyToClipboard(text)
}

async function onSaveConfig() {
  const err = await App.SaveConfig({ ...cfg })
  if (err) {
    showBanner(err)
    return
  }
  pushLog('info', '[信息]', '配置已保存')
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
        :accounts="accounts"
        @import-account-click="onImportAccountClick"
        @import-account-files="onImportAccountFiles"
        @paste-clipboard="onPasteClipboard"
        @remove-account="onRemoveAccount"
        @mark-bad="onMarkBad"
        @clear-accounts="onClearAccounts"
        @export-result="onExportResult"
        @test-account="onTestAccount"
        @copy-ck="onCopyCK"
        @save-config="onSaveConfig"
      />
      <HelpPage v-else />

      <RunParamsPanel
        v-model:threads="cfg.threads"
        v-model:interval-sec="cfg.intervalSec"
        v-model:per-account-count="cfg.perAccountCount"
        v-model:fail-switch-count="cfg.failSwitchCount"
        v-model:cycle-rounds="cfg.cycleRounds"
        v-model:round-interval-sec="cfg.roundIntervalSec"
        v-model:keyword-slots="cfg.keywordSlots"
        v-model:create-repo="cfg.createRepo"
        :content-count="contentCount"
        :running="running"
        :paused="paused"
        :round="round"
        :round-done="roundDone"
        :round-total="roundTotal"
        @select-folder="onSelectContentFolder"
        @select-files="onSelectContentFiles"
        @clear-content="onClearContent"
        @start="onStart"
        @pause="onPause"
        @resume="onResume"
        @stop="onStop"
        @save-config="onSaveConfig"
        @clear-accounts="onClearAccounts"
        @keyword-settings="onKeywordSettings"
        @switch-profile="onSwitchProfile"
        @view-results="onViewResults"
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
      :total="accounts.length"
      :success="successCount"
      :fail="failedCount"
      :pending="pendingCount"
      :elapsed="elapsed"
    />

    <ResultsModal
      v-if="showResults"
      :results="results"
      @close="showResults = false"
      @copy-all="onCopyAllResults"
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
