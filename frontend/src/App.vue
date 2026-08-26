<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { TitleBar, NavRail, LogPanel } from '@dongfang/df-ui-shell'
import ContentSettingsPage from './components/ContentSettingsPage.vue'
import ContentParamsPanel from './components/ContentParamsPanel.vue'
import PublishPage from './components/PublishPage.vue'
import HelpPage from './components/HelpPage.vue'
import ContentIcon from './icons/ContentIcon.vue'
import PublishIcon from './icons/PublishIcon.vue'
import HelpIcon from './icons/HelpIcon.vue'
import * as App from '../wailsjs/go/main/App'
import { EventsOn, WindowMinimise, WindowToggleMaximise, Quit } from '../wailsjs/runtime/runtime'

const APP_VERSION = 'v1.0.0'
const APP_NAME = 'Git MD'

const NAV_ITEMS = [
  { key: 'content', cn: '内容设置', en: 'CONTENT', icon: ContentIcon },
  { key: 'publish', cn: '发布', en: 'PUBLISH', icon: PublishIcon },
]
const NAV_BOTTOM_ITEMS = [{ key: 'help', cn: '使用说明', en: 'HELP', icon: HelpIcon }]

// 与 Go 侧 contentgen.BodyTemplateCount 对应
const BODY_TEMPLATE_COUNT = 2

const THEME_KEY = 'gitmd.theme'
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

const page = ref('content')

// 生成参数，与 Go 侧 contentgen.Options 一一对应
const opts = reactive({
  count: 5,
  keywordOrder: 'sequential',
  keywordTransform: 'none',
  shuffleParagraphs: false,
  dedupeLines: false,
  chineseOnly: false,
})

// 右侧共享配置区（发布任务参数），所有 tab 共用；目前为界面状态，落盘到 localStorage
const RUN_KEY = 'gitmd.runParams'
const RUN_DEFAULTS = {
  threads: 10,
  interval: 2,
  perAccount: 1000,
  failSwitch: 100,
  accountCycles: 1,
  roundInterval: 1,
  keywordSlots: 3,
  newRepo: false,
}
const run = reactive({ ...RUN_DEFAULTS })
try {
  const saved = JSON.parse(localStorage.getItem(RUN_KEY) || 'null')
  if (saved && typeof saved === 'object') Object.assign(run, saved)
} catch {
  /* 忽略损坏的本地配置 */
}

// 模板 tab 只编辑标题与正文模板；关键词/图片/变量/文章等词库改由「文件库」标签页
// 直接改素材目录里的 txt，不再经过这里。
const form = reactive({
  titleTemplate: '',
  bodyTemplates: Array(BODY_TEMPLATE_COUNT).fill(''),
})

// ---------- 发布页：账号队列与发布模拟 ----------
const ACCOUNTS_KEY = 'gitmd.accounts'
const accounts = ref([])
let nextAccountId = 1
try {
  const saved = JSON.parse(localStorage.getItem(ACCOUNTS_KEY) || 'null')
  if (Array.isArray(saved)) {
    accounts.value = saved
    nextAccountId = saved.reduce((m, a) => Math.max(m, a.id || 0), 0) + 1
  }
} catch {
  /* 忽略损坏的本地账号数据 */
}

const STATUS_CN = { pending: '待发', publishing: '发布中', success: '成功', failed: '失败', bad: '坏号' }

const publishing = ref(false)

function persistAccounts() {
  localStorage.setItem(ACCOUNTS_KEY, JSON.stringify(accounts.value))
}

// parseCks 把导入文本按行拆成 CK：每行一个，若含 ---- 只取分隔符前的部分，去空行。
function parseCks(text) {
  return text
    .split(/\r?\n/)
    .map((l) => l.split('----')[0].trim())
    .filter((l) => l.length > 0)
}

// addAccounts 追加账号并按 CK 去重（含已有列表与本批次内）；返回实际新增数量。
function addAccounts(cks) {
  const existing = new Set(accounts.value.map((a) => a.ck))
  let added = 0
  for (const ck of cks) {
    if (existing.has(ck)) continue
    existing.add(ck)
    accounts.value.push({
      id: nextAccountId++,
      ck,
      ua: '',
      ip: '',
      status: 'pending',
      success: 0,
      fail: 0,
      total: 0,
      bad: false,
    })
    added++
  }
  return added
}

const drafts = ref([])
const generating = ref(false)
const logs = ref([])
const autoScroll = ref(true)
const banner = ref('')

function showBanner(msg) {
  banner.value = msg
  setTimeout(() => {
    if (banner.value === msg) banner.value = ''
  }, 4000)
}

function pushLog(kind, tag, msg) {
  logs.value.push({ time: new Date().toTimeString().slice(0, 8), tag, kind, msg, highlight: false })
  if (logs.value.length > 1000) logs.value.shift()
}

// working 汇总两种忙碌态，供右侧参数面板禁用「开始工作」按钮。
const working = computed(() => generating.value || publishing.value)

const pill = computed(() => {
  if (publishing.value) return { text: '发布中…', kind: 'running' }
  if (generating.value) return { text: '生成中…', kind: 'running' }
  if (page.value === 'publish') {
    return accounts.value.length
      ? { text: `账号 ${accounts.value.length}`, kind: 'success' }
      : { text: '待机', kind: 'muted' }
  }
  if (drafts.value.length === 0) return { text: '待机', kind: 'muted' }
  return { text: `已生成 ${drafts.value.length} 篇`, kind: 'success' }
})

onMounted(async () => {
  EventsOn('gen:log', (line) => {
    logs.value.push(line)
    if (logs.value.length > 1000) logs.value.shift()
  })

  Object.assign(opts, await App.LoadConfig())
  try {
    const lib = await App.LoadContent()
    form.titleTemplate = lib.titleTemplate ?? ''
    form.bodyTemplates = Array.from({ length: BODY_TEMPLATE_COUNT }, (_, i) => lib.bodyTemplates?.[i] ?? '')
  } catch (e) {
    showBanner(String(e))
  }
})

// saveTemplates 在生成/打开目录前把模板落盘；词库不经这里，由「文件库」直接改文件。
async function saveTemplates() {
  const err = await App.SaveTemplates(form.titleTemplate, [...form.bodyTemplates])
  if (err) {
    showBanner(err)
    return false
  }
  return true
}

async function onGenerate() {
  if (generating.value) return
  generating.value = true
  try {
    if (!(await saveTemplates())) return
    drafts.value = (await App.Generate({ ...opts })) ?? []
  } catch (e) {
    showBanner(String(e))
  } finally {
    generating.value = false
  }
}

async function onExport() {
  if (drafts.value.length === 0) return
  const err = await App.ExportDrafts(drafts.value)
  if (err) showBanner(err)
}

async function onOpenDir() {
  if (!(await saveTemplates())) return
  const err = await App.OpenContentDir()
  if (err) showBanner(err)
}

async function onImportText(target) {
  try {
    const text = await App.ImportTextFile()
    if (!text) return
    if (target.startsWith('body')) {
      const next = [...form.bodyTemplates]
      next[Number(target.slice(4))] = text
      form.bodyTemplates = next
    }
  } catch (e) {
    showBanner(String(e))
  }
}

async function onCopyDraft(draft) {
  await App.CopyToClipboard(`${draft.title}\n\n${draft.body}`)
}

async function onCopyToken(token) {
  await App.CopyToClipboard(token)
  pushLog('info', '[信息]', `已复制 ${token}`)
}

async function onSaveConfig() {
  const err = await App.SaveConfig({ ...opts })
  if (err) {
    showBanner(err)
    return
  }
  localStorage.setItem(RUN_KEY, JSON.stringify({ ...run }))
  pushLog('info', '[信息]', '配置已保存')
}

// 开始工作：发布页触发发布，其它页触发内容生成。
function onStartWork() {
  if (page.value === 'publish') onPublishAll()
  else onGenerate()
}

// ---------- 发布模拟 ----------
// 目前没有真实发布后端，这里按参数模拟状态流转，方便演示与联调 UI。

const UA_POOL = ['Chrome/126 Win10', 'Chrome/125 Win11', 'Edge/126 Win10', 'Firefox/128 Win10', 'Chrome/124 Win10']
function randUA() {
  return UA_POOL[Math.floor(Math.random() * UA_POOL.length)]
}
function randIP() {
  const o = () => Math.floor(Math.random() * 254) + 1
  return `${o()}.${o()}.${o()}.${o()}`
}
function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms))
}

// publishOne 对单个账号跑一遍发布：补特征 → 发布中 → 按「每号发布」次数累加成功/失败。
// 演示上限 3 次，避免默认的 1000 次把界面卡死；每次按「发布间隔」秒节流。
async function publishOne(a) {
  if (a.bad || a.status === 'bad') return
  if (!a.ua) a.ua = randUA()
  if (!a.ip) a.ip = randIP()
  a.status = 'publishing'
  const attempts = Math.min(Math.max(run.perAccount, 1), 3)
  let lastOk = false
  for (let i = 0; i < attempts; i++) {
    await sleep(Math.max(run.interval, 0) * 1000 || 200)
    lastOk = Math.random() < 0.78
    if (lastOk) a.success++
    else a.fail++
    a.total++
  }
  a.status = lastOk ? 'success' : 'failed'
}

// onPublishAll 用线程数量控制并发，逐个消费待发账号。
async function onPublishAll() {
  if (working.value) return
  const pending = accounts.value.filter((a) => a.status === 'pending' && !a.bad)
  if (pending.length === 0) {
    pushLog('info', '[信息]', '没有待发账号')
    return
  }
  publishing.value = true
  pushLog('start', '[开始]', `开始发布，待发 ${pending.length} 个 · 线程 ${run.threads}（模拟）`)
  try {
    let idx = 0
    const worker = async () => {
      while (idx < pending.length) {
        const a = pending[idx++]
        await publishOne(a)
        pushLog(a.status === 'success' ? 'success' : 'failure', a.status === 'success' ? '[成功]' : '[失败]', shortCk(a.ck))
      }
    }
    const n = Math.min(Math.max(run.threads, 1), pending.length)
    await Promise.all(Array.from({ length: n }, worker))
    const ok = pending.filter((a) => a.status === 'success').length
    pushLog('info', '[信息]', `发布完成：成功 ${ok}，失败 ${pending.length - ok}`)
    persistAccounts()
  } finally {
    publishing.value = false
  }
}

function shortCk(ck) {
  return ck.length > 16 ? ck.slice(0, 16) + '…' : ck
}

// ---------- 发布页动作 ----------
async function onImportAccountsFile() {
  try {
    const text = await App.ImportTextFile()
    if (!text) return
    const added = addAccounts(parseCks(text))
    pushLog('info', '[信息]', `导入账号：新增 ${added} 个`)
  } catch (e) {
    showBanner(String(e))
  }
}

async function onPasteClipboard() {
  const text = await App.ClipboardGetText()
  if (!text || !text.trim()) {
    pushLog('retry', '[提示]', '剪贴板为空')
    return
  }
  const added = addAccounts(parseCks(text))
  pushLog('info', '[信息]', `剪贴板导入：新增 ${added} 个`)
}

function onImportRaw(text) {
  const added = addAccounts(parseCks(text))
  pushLog('info', '[信息]', `拖入导入：新增 ${added} 个`)
}

async function onExportResults() {
  if (accounts.value.length === 0) {
    pushLog('info', '[信息]', '没有账号可导出')
    return
  }
  const header = ['序号', 'CK', 'UA', 'IP', '状态', '成功', '失败', '总数'].join('\t')
  const body = accounts.value.map((a, i) =>
    [i + 1, a.ck, a.ua || '', a.ip || '', STATUS_CN[a.status] ?? a.status, a.success, a.fail, a.total].join('\t'),
  )
  const err = await App.SaveTextFile('publish-results.txt', [header, ...body].join('\n'))
  if (err) showBanner(err)
}

function onSaveAccounts() {
  persistAccounts()
  pushLog('info', '[信息]', `发布配置已保存（${accounts.value.length} 个账号）`)
}

// 清空账号：发布页与右侧面板共用，直接清空真实列表。
function onClearAccounts() {
  const n = accounts.value.length
  accounts.value = []
  persistAccounts()
  pushLog('info', '[信息]', `已清空 ${n} 个账号`)
}

async function onCopyCk(a) {
  await App.CopyToClipboard(a.ck)
  pushLog('info', '[信息]', '已复制 CK')
}

async function onTestAccount(a) {
  if (working.value) return
  publishing.value = true
  try {
    pushLog('start', '[开始]', `单独测试 ${shortCk(a.ck)}（模拟）`)
    await publishOne(a)
    pushLog(a.status === 'success' ? 'success' : 'failure', '[结果]', `${shortCk(a.ck)} → ${STATUS_CN[a.status] ?? a.status}`)
    persistAccounts()
  } finally {
    publishing.value = false
  }
}

function onRemoveAccount(a) {
  accounts.value = accounts.value.filter((x) => x.id !== a.id)
  persistAccounts()
}

function onMarkBad(a) {
  a.bad = true
  a.status = 'bad'
  persistAccounts()
  pushLog('retry', '[提示]', `已标记坏号 ${shortCk(a.ck)}`)
}

// 关键词设置：跳回内容页。
function onKeywordSettings() {
  page.value = 'content'
  pushLog('info', '[信息]', '请在左侧「变量设置」中配置关键词库')
}

// 换号特征：给所有非坏号账号重新分配一套 UA / IP 特征。
function onAccountFeature() {
  let n = 0
  for (const a of accounts.value) {
    if (a.bad) continue
    a.ua = randUA()
    a.ip = randIP()
    n++
  }
  persistAccounts()
  pushLog('info', '[信息]', `已为 ${n} 个账号换号特征`)
}

function onViewLinks() {
  pushLog('info', '[信息]', '查看链接（功能待接入）')
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
    <TitleBar
      :theme="theme"
      :version="APP_VERSION"
      :pill-text="pill.text"
      :pill-kind="pill.kind"
      :app-name="APP_NAME"
      logo-text="M"
      @toggle-theme="toggleTheme"
      @minimize="WindowMinimise"
      @maximize="WindowToggleMaximise"
      @close="Quit"
    />

    <div v-if="banner" class="banner">{{ banner }}</div>

    <div class="body">
      <NavRail :page="page" :items="NAV_ITEMS" :bottom-items="NAV_BOTTOM_ITEMS" @navigate="(p) => (page = p)" />

      <HelpPage v-if="page === 'help'" />
      <PublishPage
        v-else-if="page === 'publish'"
        :accounts="accounts"
        @import-file="onImportAccountsFile"
        @paste-clipboard="onPasteClipboard"
        @import-raw="onImportRaw"
        @export-results="onExportResults"
        @save-config="onSaveAccounts"
        @clear-accounts="onClearAccounts"
        @copy-ck="onCopyCk"
        @test-account="onTestAccount"
        @remove-account="onRemoveAccount"
        @mark-bad="onMarkBad"
      />
      <ContentSettingsPage
        v-else
        v-model:title-template="form.titleTemplate"
        v-model:body-templates="form.bodyTemplates"
        v-model:keyword-order="opts.keywordOrder"
        v-model:keyword-transform="opts.keywordTransform"
        v-model:shuffle-paragraphs="opts.shuffleParagraphs"
        :drafts="drafts"
        @import-text="onImportText"
        @copy-draft="onCopyDraft"
        @copy-token="onCopyToken"
        @open-dir="onOpenDir"
      />

      <ContentParamsPanel
        v-model:threads="run.threads"
        v-model:interval="run.interval"
        v-model:per-account="run.perAccount"
        v-model:fail-switch="run.failSwitch"
        v-model:account-cycles="run.accountCycles"
        v-model:round-interval="run.roundInterval"
        v-model:keyword-slots="run.keywordSlots"
        v-model:new-repo="run.newRepo"
        :working="working"
        @keyword-settings="onKeywordSettings"
        @save-config="onSaveConfig"
        @clear-accounts="onClearAccounts"
        @account-feature="onAccountFeature"
        @view-links="onViewLinks"
        @start-work="onStartWork"
      />
    </div>

    <LogPanel
      :lines="logs"
      v-model:auto-scroll="autoScroll"
      @copy="onCopyLog"
      @export="onExportLog"
      @clear="onClearLog"
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
