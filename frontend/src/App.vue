<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { TitleBar, NavRail, LogPanel } from '@dongfang/df-ui-shell'
import ContentSettingsPage from './components/ContentSettingsPage.vue'
import ContentParamsPanel from './components/ContentParamsPanel.vue'
import HelpPage from './components/HelpPage.vue'
import ContentIcon from './icons/ContentIcon.vue'
import HelpIcon from './icons/HelpIcon.vue'
import * as App from '../wailsjs/go/main/App'
import { EventsOn, WindowMinimise, WindowToggleMaximise, Quit } from '../wailsjs/runtime/runtime'

const APP_VERSION = 'v1.0.0'
const APP_NAME = 'Git MD'

const NAV_ITEMS = [{ key: 'content', cn: '内容设置', en: 'CONTENT', icon: ContentIcon }]
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

const pill = computed(() => {
  if (generating.value) return { text: '生成中…', kind: 'running' }
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

// 开始工作：先落盘内容，再触发生成/发布流程
function onStartWork() {
  onGenerate()
}

// 以下按钮对应参考 UI，功能尚未接入后端，先给出提示占位
function onKeywordSettings() {
  page.value = 'content'
  pushLog('info', '[信息]', '请在左侧「变量设置」中配置关键词库')
}
function onClearAccounts() {
  pushLog('info', '[信息]', '清空账号（功能待接入）')
}
function onAccountFeature() {
  pushLog('info', '[信息]', '换号特征（功能待接入）')
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
        :working="generating"
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
