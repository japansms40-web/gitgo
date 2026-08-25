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

// 与 Go 侧 contentgen.VarBankCount / BodyTemplateCount 对应
const VAR_BANK_COUNT = 5
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

// 素材在界面上一律按纯文本编辑，存盘时才切成一行一条
const form = reactive({
  titleTemplate: '',
  bodyTemplates: Array(BODY_TEMPLATE_COUNT).fill(''),
  keywordsText: '',
  imagesText: '',
  varTexts: Array(VAR_BANK_COUNT).fill(''),
})

// 文章库是用户往素材目录里丢文件，界面只读不改，所以单独存着回写用
const articles = ref([])

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

function splitLines(text) {
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line !== '')
}

function toLibrary() {
  return {
    titleTemplate: form.titleTemplate,
    bodyTemplates: [...form.bodyTemplates],
    keywords: splitLines(form.keywordsText),
    vars: form.varTexts.map(splitLines),
    images: splitLines(form.imagesText),
    articles: articles.value,
  }
}

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
    form.keywordsText = (lib.keywords ?? []).join('\n')
    form.imagesText = (lib.images ?? []).join('\n')
    form.varTexts = Array.from({ length: VAR_BANK_COUNT }, (_, i) => (lib.vars?.[i] ?? []).join('\n'))
    articles.value = lib.articles ?? []
  } catch (e) {
    showBanner(String(e))
  }
})

// saveContent 在生成/打开目录前先落盘，保证界面上看到的和 txt 里存的是同一份。
async function saveContent() {
  const err = await App.SaveContent(toLibrary())
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
    if (!(await saveContent())) return
    drafts.value = (await App.Generate({ ...opts })) ?? []
    // 用户可能刚往「文章库」目录丢了文件，顺手刷新一下计数
    articles.value = (await App.LoadContent()).articles ?? []
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
  if (!(await saveContent())) return
  const err = await App.OpenContentDir()
  if (err) showBanner(err)
}

async function onImportText(target) {
  try {
    const text = await App.ImportTextFile()
    if (!text) return
    if (target === 'title') form.titleTemplate = text
    else if (target === 'keywords') form.keywordsText = text
    else if (target === 'images') form.imagesText = text
    else if (target.startsWith('body')) {
      const next = [...form.bodyTemplates]
      next[Number(target.slice(4))] = text
      form.bodyTemplates = next
    } else if (target.startsWith('var')) {
      const next = [...form.varTexts]
      next[Number(target.slice(3))] = text
      form.varTexts = next
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
        v-model:keywords-text="form.keywordsText"
        v-model:images-text="form.imagesText"
        v-model:var-texts="form.varTexts"
        v-model:keyword-order="opts.keywordOrder"
        v-model:keyword-transform="opts.keywordTransform"
        v-model:shuffle-paragraphs="opts.shuffleParagraphs"
        :article-count="articles.length"
        :drafts="drafts"
        @import-text="onImportText"
        @copy-draft="onCopyDraft"
        @copy-token="onCopyToken"
        @open-dir="onOpenDir"
      />

      <ContentParamsPanel
        v-model:count="opts.count"
        v-model:dedupe-lines="opts.dedupeLines"
        v-model:chinese-only="opts.chineseOnly"
        :draft-count="drafts.length"
        :generating="generating"
        @generate="onGenerate"
        @export="onExport"
        @save-config="onSaveConfig"
        @open-dir="onOpenDir"
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
