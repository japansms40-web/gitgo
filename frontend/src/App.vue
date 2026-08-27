<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { TitleBar, NavRail } from '@dongfang/df-ui-shell'
import LogPanel from './components/LogPanel.vue'
import ContentSettingsPage from './components/ContentSettingsPage.vue'
import ContentParamsPanel from './components/ContentParamsPanel.vue'
import PublishPage from './components/PublishPage.vue'
import ProxyPage from './components/ProxyPage.vue'
import HelpPage from './components/HelpPage.vue'
import ContentIcon from './icons/ContentIcon.vue'
import PublishIcon from './icons/PublishIcon.vue'
import ProxyIcon from './icons/ProxyIcon.vue'
import HelpIcon from './icons/HelpIcon.vue'
import * as App from '../wailsjs/go/main/App'
import { EventsOn, WindowMinimise, WindowToggleMaximise, Quit } from '../wailsjs/runtime/runtime'

const APP_VERSION = 'v1.2.0'
const APP_NAME = 'Git MD'

const NAV_ITEMS = [
  { key: 'publish', cn: '发布', en: 'PUBLISH', icon: PublishIcon },
  { key: 'content', cn: '内容设置', en: 'CONTENT', icon: ContentIcon },
  { key: 'proxy', cn: '代理', en: 'PROXY', icon: ProxyIcon },
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

// ---------- 代理设置（单个全局代理，落盘 localStorage）----------
const PROXY_KEY = 'gitmd.proxy'
const proxy = reactive({ enabled: false, url: '' })
try {
  const saved = JSON.parse(localStorage.getItem(PROXY_KEY) || 'null')
  if (saved && typeof saved === 'object') Object.assign(proxy, saved)
} catch {
  /* 忽略损坏的本地代理配置 */
}
const proxyTesting = ref(false)
const proxyTestResult = ref(null)

// ---------- 发布链接（输出到文件库 查看链接.txt）----------
const LINKS_FILE = '查看链接.txt'
const links = ref([]) // {repo, file, url}
// openLibFile 是给文件库的「打开指定文件」信号：{ path, n }，n 变化即重新触发。
const openLibFile = ref(null)

function onSaveProxy() {
  localStorage.setItem(PROXY_KEY, JSON.stringify({ ...proxy }))
  pushLog('info', '[信息]', proxy.enabled ? `代理已保存并启用：${proxy.url || '(空)'}` : '代理已保存（未启用）')
}

// onTestProxy 走 Go 侧 TestProxy 真拨测 github.com。
async function onTestProxy() {
  if (proxyTesting.value) return
  proxyTesting.value = true
  proxyTestResult.value = null
  try {
    proxyTestResult.value = await App.TestProxy(proxy.url)
  } catch (e) {
    proxyTestResult.value = { ok: false, message: '拨测异常：' + String(e) }
  } finally {
    proxyTesting.value = false
  }
}

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
const LOG_MAX = 1000 // 日志面板只保留最近 N 行（全量历史在磁盘日志文件里）
let nextLogId = 1 // 每行日志的稳定 key，供 LogPanel 复用 DOM、避免 shift 触发全表重排
const autoScroll = ref(true)
const banner = ref('')

function showBanner(msg) {
  banner.value = msg
  setTimeout(() => {
    if (banner.value === msg) banner.value = ''
  }, 4000)
}

function pushLog(kind, tag, msg) {
  logs.value.push({ _id: nextLogId++, time: new Date().toTimeString().slice(0, 8), tag, kind, msg, highlight: false })
  if (logs.value.length > LOG_MAX) logs.value.splice(0, logs.value.length - LOG_MAX)
  // 前端产生的日志也落盘到当天日志文件（后端日志由后端直接写，不经这里，避免重复）
  App.AppendLog(tag, msg)
}

// onViewLogs 打开本地日志目录（按日期分文件），供直接查看日志文件。
async function onViewLogs() {
  const err = await App.OpenLogsDir()
  if (err) showBanner(err)
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
  // 后端按 ~100ms 微批合并，载荷为数组；遍历入库。稳定 _id 让 LogPanel 复用 DOM。
  EventsOn('gen:log', (lines) => {
    for (const line of lines) {
      line._id = nextLogId++
      logs.value.push(line)
    }
    if (logs.value.length > LOG_MAX) logs.value.splice(0, logs.value.length - LOG_MAX)
  })

  // 发布进度：单账号状态更新（数组，同窗口每号只带最新累计值）
  EventsOn('publish:account', (updates) => {
    for (const u of updates) {
      const a = accounts.value.find((x) => x.id === u.id)
      if (!a) continue
      a.status = u.status
      a.success = u.success
      a.fail = u.fail
      if (u.status === 'bad') a.bad = true
    }
  })
  // 发布成功一篇：收集链接（数组），供「查看链接」输出到文件
  EventsOn('publish:link', (arr) => {
    for (const l of arr) links.value.push({ repo: l.repo, file: l.file, url: l.url })
    if (links.value.length > 20000) links.value.splice(0, links.value.length - 20000)
  })
  // 发布任务结束（链接已由后端边发边增量写入 查看链接.txt）
  EventsOn('publish:done', () => {
    publishing.value = false
    persistAccounts()
    pushLog('info', '[信息]', `发布任务结束，共 ${links.value.length} 条链接`)
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

// 开始工作：发布页触发/停止真实发布，其它页触发内容生成。
function onStartWork() {
  if (page.value === 'publish') {
    if (publishing.value) onStopPublish()
    else onPublishAll()
  } else {
    if (!generating.value) onGenerate()
  }
}

// onStopPublish 请求后端取消发布任务。
async function onStopPublish() {
  await App.StopPublish()
  pushLog('retry', '[提示]', '正在停止发布…')
}

// ---------- 账号特征（换号特征用）----------

const UA_POOL = ['Chrome/126 Win10', 'Chrome/125 Win11', 'Edge/126 Win10', 'Firefox/128 Win10', 'Chrome/124 Win10']
function randUA() {
  return UA_POOL[Math.floor(Math.random() * UA_POOL.length)]
}
function randIP() {
  const o = () => Math.floor(Math.random() * 254) + 1
  return `${o()}.${o()}.${o()}.${o()}`
}

// onPublishAll 启动真实批量发布：账号 + 配置 + 内容选项交给后端 worker-pool 引擎，
// 进度经 publish:account / publish:done 事件回写。StartPublish 立即返回，任务在后台跑。
async function onPublishAll() {
  if (publishing.value) return
  const targets = accounts.value.filter((a) => a.ck && !a.bad)
  if (targets.length === 0) {
    pushLog('info', '[信息]', '没有可发布的账号')
    return
  }
  // 先把模板落盘，保证后端读到最新内容
  if (!(await saveTemplates())) return
  links.value = [] // 新一轮发布清空旧链接
  publishing.value = true
  const cfg = {
    threads: run.threads,
    interval: run.interval,
    perAccount: run.perAccount,
    failSwitch: run.failSwitch,
    accountCycles: run.accountCycles,
    roundInterval: run.roundInterval,
    proxyUrl: proxy.enabled ? proxy.url : '',
  }
  const payload = accounts.value.map((a) => ({ id: a.id, ck: a.ck }))
  try {
    const err = await App.StartPublish(payload, cfg, { ...opts })
    if (err) {
      publishing.value = false
      showBanner(err)
      pushLog('failure', '[失败]', err)
    }
  } catch (e) {
    publishing.value = false
    showBanner(String(e))
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

// onTestAccount 真实验活：用账号 CK + 保存的代理调 Go 侧 CheckAccount（走真实 ListRepos）。
async function onTestAccount(a) {
  if (working.value) return
  publishing.value = true
  const prev = a.status
  try {
    pushLog('start', '[开始]', `验活 ${shortCk(a.ck)}${proxy.enabled ? ' · 走代理' : ''}`)
    a.status = 'publishing'
    const res = await App.CheckAccount(a.ck, proxy.enabled ? proxy.url : '')
    if (res.ok) {
      a.status = 'success'
      a.bad = false
      pushLog('success', '[活号]', `${shortCk(a.ck)} · 仓库 ${res.repoCount} 个`)
    } else if (res.bad) {
      a.status = 'bad'
      a.bad = true
      pushLog('failure', '[坏号]', `${shortCk(a.ck)} · ${res.message}`)
    } else {
      a.status = prev === 'publishing' ? 'pending' : prev
      pushLog('failure', '[失败]', `${shortCk(a.ck)} · ${res.message}`)
    }
    persistAccounts()
  } catch (e) {
    a.status = prev === 'publishing' ? 'pending' : prev
    pushLog('failure', '[失败]', `${shortCk(a.ck)} · ${String(e)}`)
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

// onViewLinks 跳到「内容设置 → 文件库」并打开 查看链接.txt。
// 链接在发布过程中已由后端增量写入该文件；这里只确保文件存在（从没发过则建空）再跳转。
async function onViewLinks() {
  const err = await App.EnsureLinksFile()
  if (err) {
    showBanner(err)
    return
  }
  page.value = 'content'
  // 变更 n 触发文件库重新加载并选中该文件（Date.now 仅作触发用，不参与逻辑）。
  // tail：链接文件可达数十万行，预览只加载尾部最新链接，避免整份加载卡死。
  openLibFile.value = { path: LINKS_FILE, n: Date.now(), tail: true }
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
      <ProxyPage
        v-else-if="page === 'proxy'"
        v-model:url="proxy.url"
        v-model:enabled="proxy.enabled"
        :testing="proxyTesting"
        :test-result="proxyTestResult"
        @test="onTestProxy"
        @save="onSaveProxy"
      />
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
        :open-file="openLibFile"
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
      @view-logs="onViewLogs"
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
