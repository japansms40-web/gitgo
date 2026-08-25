<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  titleTemplate: { type: String, required: true },
  bodyTemplates: { type: Array, required: true }, // A / B 两套正文模板
  keywordsText: { type: String, required: true },
  imagesText: { type: String, required: true },
  articleCount: { type: Number, required: true },
  varTexts: { type: Array, required: true }, // 5 段纯文本，一行一条
  keywordOrder: { type: String, required: true },
  keywordTransform: { type: String, required: true },
  shuffleParagraphs: { type: Boolean, required: true },
  drafts: { type: Array, required: true }, // {title, body}
})
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
])

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

const activeTab = ref('template')
const tabs = computed(() => [
  { key: 'template', label: '模板' },
  { key: 'bank', label: '变量设置' },
  { key: 'preview', label: `预览 ${props.drafts.length}` },
])

const expanded = ref(-1)
function toggle(index) {
  expanded.value = expanded.value === index ? -1 : index
}

function lineCount(text) {
  return text.split('\n').filter((l) => l.trim() !== '').length
}

function updateBody(index, value) {
  const next = [...props.bodyTemplates]
  next[index] = value
  emit('update:bodyTemplates', next)
}

function updateVar(index, value) {
  const next = [...props.varTexts]
  next[index] = value
  emit('update:varTexts', next)
}

function summary(bodyText) {
  const first = bodyText.split('\n').find((l) => l.trim() !== '') || ''
  return first.length > 60 ? first.slice(0, 60) + '…' : first
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">
        <span class="page-title-main">内容设置 Content</span>
        <span class="page-title-sub">标题与正文模板、标签变量</span>
      </div>
      <div class="spacer" />
      <div class="tabs">
        <button
          v-for="t in tabs"
          :key="t.key"
          class="tab"
          :class="{ active: activeTab === t.key }"
          @click="activeTab = t.key"
        >{{ t.label }}</button>
      </div>
    </div>

    <!-- 模板 -->
    <div v-if="activeTab === 'template'" class="page-body two-col">
      <div class="col-main">
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
        </div>

        <fieldset class="box">
          <legend>标题模板 Title</legend>
          <input
            class="input"
            placeholder="例如：{关键词}-{英文=5}{小写=3}"
            :value="titleTemplate"
            @input="emit('update:titleTemplate', $event.target.value)"
          />
        </fieldset>

        <fieldset class="box box-grow">
          <legend>正文模板 Body · 支持右侧全部标签</legend>

          <div class="option-row">
            <label class="checkbox-row">
              <input
                type="checkbox"
                class="sr-only"
                :checked="shuffleParagraphs"
                @change="emit('update:shuffleParagraphs', $event.target.checked)"
              />
              <span class="checkbox-box" :class="{ checked: shuffleParagraphs }">
                <svg v-if="shuffleParagraphs" width="10" height="10" viewBox="0 0 16 16" fill="none" stroke="#fff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M3 8.2l3.3 3.3L13 4.5" />
                </svg>
              </span>
              随机打乱段落
            </label>
            <div class="spacer" />
            <span class="option-note">模板 A / B 随机选用</span>
          </div>

          <div v-for="(tpl, i) in bodyTemplates" :key="i" class="body-slot">
            <div class="body-head">
              <span class="body-label mono">{{ i === 0 ? 'A' : 'B' }}</span>
              <div class="spacer" />
              <button class="link-action" @click="emit('import-text', 'body' + i)">从文件导入</button>
            </div>
            <textarea
              class="input body-input"
              :placeholder="i === 0 ? '正文内容，可混用标签与固定文字' : '第二套模板，留空则只用 A'"
              :value="tpl"
              @input="updateBody(i, $event.target.value)"
            />
          </div>
        </fieldset>
      </div>

      <fieldset class="box col-side">
        <legend>标签格式 · 点击复制</legend>
        <div class="token-cols">
          <div v-for="(col, ci) in TOKEN_COLUMNS" :key="ci" class="token-col">
            <button
              v-for="tk in col"
              :key="tk.t"
              class="token-chip mono"
              :title="tk.d"
              @click="emit('copy-token', tk.t)"
            >{{ tk.t }}</button>
          </div>
        </div>
        <div class="token-note">= 后的数字为随机长度，可自行修改；点击即写入剪贴板。同一个标签出现多次会各自重新抽取。</div>
      </fieldset>
    </div>

    <!-- 变量设置 -->
    <div v-else-if="activeTab === 'bank'" class="page-body">
      <fieldset class="box">
        <legend>关键词库</legend>
        <div class="field-head">
          <span class="field-count mono">{{ lineCount(keywordsText) }} 条</span>
          <div class="spacer" />
          <button class="link-action" @click="emit('import-text', 'keywords')">从文件导入</button>
        </div>
        <textarea
          class="input"
          rows="5"
          placeholder="一行一个关键词"
          :value="keywordsText"
          @input="emit('update:keywordsText', $event.target.value)"
        />
      </fieldset>

      <fieldset class="box">
        <legend>图片库</legend>
        <div class="field-head">
          <code class="mono field-token">{{ '{图片}' }}</code>
          <span class="field-count mono">{{ lineCount(imagesText) }} 条</span>
          <div class="spacer" />
          <button class="link-action" @click="emit('import-text', 'images')">从文件导入</button>
        </div>
        <textarea
          class="input"
          rows="3"
          placeholder="一行一条图片地址；写成 Markdown 或 HTML 也可以，会原样使用"
          :value="imagesText"
          @input="emit('update:imagesText', $event.target.value)"
        />
      </fieldset>

      <fieldset class="box">
        <legend>文章库</legend>
        <div class="field-head">
          <code class="mono field-token">{{ '{文章} {文章名}' }}</code>
          <span class="field-count mono">{{ articleCount }} 篇</span>
          <div class="spacer" />
          <button class="link-action" @click="emit('open-dir')">打开目录</button>
        </div>
        <div class="hint">
          把 .txt / .md 文件放进素材目录下的「文章库」文件夹即可。{{ '{文章}' }} 取正文，{{ '{文章名}' }} 取文件名，
          同一篇草稿里这两个标签取自同一份素材。
        </div>
      </fieldset>

      <fieldset v-for="(text, i) in varTexts" :key="i" class="box">
        <legend>变量库 {{ i + 1 }}</legend>
        <div class="field-head">
          <code class="mono field-token">{{ '{变量' + (i + 1) + '}' }}</code>
          <span class="field-count mono">{{ lineCount(text) }} 条</span>
          <div class="spacer" />
          <button class="link-action" @click="emit('import-text', 'var' + i)">从文件导入</button>
        </div>
        <textarea
          class="input"
          rows="4"
          placeholder="一行一条候选内容，生成时随机抽一行"
          :value="text"
          @input="updateVar(i, $event.target.value)"
        />
      </fieldset>
    </div>

    <!-- 预览 -->
    <div v-else class="page-body">
      <div v-if="drafts.length === 0" class="empty">还没有生成结果，点右侧「生成」试试</div>
      <div v-for="(d, i) in drafts" :key="i" class="draft" :class="{ open: expanded === i }">
        <div class="draft-head" @click="toggle(i)">
          <span class="draft-index mono muted">{{ i + 1 }}</span>
          <div class="draft-text">
            <span class="draft-title">{{ d.title }}</span>
            <span v-if="expanded !== i" class="draft-summary">{{ summary(d.body) }}</span>
          </div>
          <button class="link-action" @click.stop="emit('copy-draft', d)">复制</button>
        </div>
        <pre v-if="expanded === i" class="draft-body">{{ d.body }}</pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: var(--surface);
}
.page-header {
  height: 54px;
  flex: 0 0 54px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 18px;
  border-bottom: 1px solid var(--border);
}
.page-title {
  display: flex;
  flex-direction: column;
  line-height: 1.25;
}
.page-title-main {
  font-size: 15px;
  font-weight: 700;
}
.page-title-sub {
  font-size: 11px;
  color: var(--muted);
}
.spacer {
  flex: 1;
}
.tabs {
  display: flex;
  gap: 6px;
}
.tab {
  height: 32px;
  padding: 0 12px;
  border-radius: 16px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  font-size: 12.5px;
  cursor: pointer;
  white-space: nowrap;
}
.tab.active {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-weak);
}
.page-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 14px 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.two-col {
  flex-direction: row;
  gap: 14px;
}
.col-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.col-side {
  flex: 0 0 268px;
  align-self: flex-start;
}

/* fieldset + legend 还原截图里带标题的分组框 */
.box {
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 10px 12px 12px;
  margin: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.box legend {
  padding: 0 6px;
  font-size: 11.5px;
  color: var(--muted);
}
.box-grow {
  flex: 1;
  min-height: 0;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}
.toolbar .select {
  flex: 1;
  min-width: 0;
}
.input {
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  background: var(--surface);
  color: var(--text);
  font-size: 12.5px;
  font-family: inherit;
  padding: 8px 10px;
  outline: none;
  line-height: 1.6;
  resize: vertical;
}
.input:focus {
  border-color: var(--accent);
}
.select {
  height: 32px;
  padding: 0 8px;
  cursor: pointer;
}
.btn-primary {
  height: 32px;
  padding: 0 14px;
  border-radius: 5px;
  border: none;
  background: var(--accent);
  color: #fff;
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
}

.option-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.option-note {
  font-size: 11.5px;
  color: var(--muted);
}
.checkbox-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  color: var(--text);
  cursor: pointer;
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
.checkbox-box {
  flex: 0 0 15px;
  width: 15px;
  height: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-strong);
  border-radius: 3px;
  background: var(--surface);
}
.checkbox-box.checked {
  background: var(--accent);
  border-color: var(--accent);
}

.body-slot {
  flex: 1;
  min-height: 96px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.body-head {
  display: flex;
  align-items: center;
  gap: 8px;
}
.body-label {
  font-size: 11.5px;
  font-weight: 600;
  color: var(--accent);
}
.body-input {
  flex: 1;
  min-height: 80px;
  resize: none;
}

.field-head {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
}
.field-token {
  color: var(--accent);
  font-size: 11.5px;
}
.field-count {
  color: var(--muted);
  font-size: 11.5px;
}
.link-action {
  border: none;
  background: none;
  padding: 0;
  color: var(--accent);
  font-size: 12px;
  cursor: pointer;
}
.hint {
  font-size: 11.5px;
  line-height: 1.7;
  color: var(--muted);
}

.token-cols {
  display: flex;
  gap: 8px;
}
.token-col {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.token-chip {
  height: 34px;
  padding: 0 4px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface-2);
  color: var(--text);
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.token-chip:hover {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-weak);
}
.token-note {
  margin-top: 2px;
  font-size: 11.5px;
  line-height: 1.7;
  color: var(--muted);
}

.empty {
  padding: 40px 0;
  text-align: center;
  color: var(--muted);
  font-size: 13px;
}
.draft {
  border: 1px solid var(--border);
  border-radius: 6px;
  overflow: hidden;
  flex: 0 0 auto;
}
.draft.open {
  border-color: var(--accent);
}
.draft-head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  cursor: pointer;
}
.draft-head:hover {
  background: var(--surface-2);
}
.draft-index {
  flex: 0 0 24px;
  font-size: 11.5px;
}
.muted {
  color: var(--muted);
}
.draft-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.draft-title {
  font-size: 12.5px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.draft-summary {
  font-size: 11.5px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.draft-body {
  margin: 0;
  padding: 12px;
  border-top: 1px solid var(--border);
  background: var(--surface-2);
  font-family: inherit;
  font-size: 12.5px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 320px;
  overflow: auto;
}
</style>
