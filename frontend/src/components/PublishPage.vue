<script setup>
import { computed } from 'vue'

const props = defineProps({
  token: { type: String, required: true },
  owner: { type: String, required: true },
  repo: { type: String, required: true },
  branch: { type: String, required: true },
  dir: { type: String, required: true },
  tokenStatus: { type: String, required: true },
  queue: { type: Array, required: true }, // {title, repoPath, status}
})
const emit = defineEmits([
  'update:token',
  'update:owner',
  'update:repo',
  'update:branch',
  'update:dir',
  'validate-token',
  'select-folder',
  'select-files',
  'clear-queue',
])

const STATUS_STYLE = {
  待发布: { bg: 'transparent', fg: 'var(--muted)' },
  发布中: { bg: 'rgba(154,103,0,.16)', fg: 'var(--warn)' },
  成功: { bg: 'rgba(26,127,55,.14)', fg: 'var(--ok)' },
}
function statusStyle(status) {
  if (STATUS_STYLE[status]) return STATUS_STYLE[status]
  if (status.startsWith('失败')) return { bg: 'rgba(207,34,46,.12)', fg: 'var(--err)' }
  if (status.startsWith('重试中')) return { bg: 'rgba(154,103,0,.16)', fg: 'var(--warn)' }
  return STATUS_STYLE['待发布']
}

const done = computed(() => props.queue.filter((q) => q.status === '成功' || q.status.startsWith('失败')).length)
const progressPercent = computed(() => (props.queue.length === 0 ? 0 : Math.round((done.value / props.queue.length) * 100)))
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">
        <span class="page-title-main">发布设置</span>
        <span class="page-title-sub">Token / 仓库 / 分支等基础参数</span>
      </div>
      <div class="spacer" />
      <div class="page-actions">
        <button class="btn-outline" @click="emit('select-folder')">选择文件夹</button>
        <button class="btn-outline" @click="emit('select-files')">添加文件</button>
        <button class="btn-outline" @click="emit('clear-queue')">清空</button>
      </div>
    </div>

    <div class="page-body">
      <div class="settings-box">
        <div class="field">
          <label>Token</label>
          <div class="field-input-row">
            <input
              class="field-input"
              type="password"
              :value="token"
              @input="emit('update:token', $event.target.value)"
            />
            <button class="btn-outline btn-nowrap" @click="emit('validate-token')">验证 Token</button>
          </div>
        </div>
        <div v-if="tokenStatus" class="token-status">{{ tokenStatus }}</div>

        <div class="field">
          <label>Owner</label>
          <input class="field-input" :value="owner" @input="emit('update:owner', $event.target.value)" />
        </div>
        <div class="field">
          <label>Repo</label>
          <input class="field-input" :value="repo" @input="emit('update:repo', $event.target.value)" />
        </div>
        <div class="field">
          <label>分支</label>
          <input class="field-input" :value="branch" @input="emit('update:branch', $event.target.value)" />
        </div>
        <div class="field">
          <label>目标目录</label>
          <input class="field-input" :value="dir" @input="emit('update:dir', $event.target.value)" />
        </div>
      </div>

      <div class="table-box">
        <div class="table-header">
          <div>序号</div>
          <div>文件名</div>
          <div>仓库路径</div>
          <div class="align-right">状态</div>
        </div>
        <div v-for="(item, i) in queue" :key="i" class="table-row">
          <div class="mono muted">{{ i + 1 }}</div>
          <div class="ellipsis">{{ item.title }}</div>
          <div class="ellipsis muted">{{ item.repoPath }}</div>
          <div class="align-right">
            <span
              class="status-badge"
              :style="{ background: statusStyle(item.status).bg, color: statusStyle(item.status).fg }"
            >{{ item.status }}</span>
          </div>
        </div>
      </div>

      <div class="progress-row">
        <span>已发布 {{ done }} / 队列 {{ queue.length }} 篇</span>
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: progressPercent + '%' }" />
        </div>
        <span class="mono">{{ progressPercent }}%</span>
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
.page-actions {
  display: flex;
  gap: 7px;
}
.btn-outline {
  height: 30px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  border-radius: 5px;
  border: 1px solid var(--border-strong);
  background: var(--surface);
  font-size: 12.5px;
  cursor: pointer;
  color: var(--text);
}
.btn-outline:hover {
  border-color: var(--accent);
  color: var(--accent);
}
.btn-nowrap {
  white-space: nowrap;
}
.page-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.settings-box {
  border: 1px solid var(--border);
  border-radius: 7px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.field {
  display: flex;
  align-items: center;
  gap: 16px;
}
.field > label {
  flex: 0 0 90px;
  font-size: 13px;
  font-weight: 600;
  text-align: right;
}
.field-input-row {
  flex: 1;
  display: flex;
  gap: 8px;
}
.field-input {
  flex: 1;
  height: 34px;
  border: 1px solid var(--border-strong);
  border-radius: 5px;
  background: var(--surface);
  padding: 0 11px;
  font-size: 13px;
  color: var(--text);
  font-family: inherit;
}
.field-input:focus {
  outline: none;
  border-color: var(--accent);
}
.token-status {
  margin-left: 106px;
  font-size: 12px;
  color: var(--muted);
}
.table-box {
  border: 1px solid var(--border);
  border-radius: 7px;
  overflow: hidden;
}
.table-header,
.table-row {
  display: grid;
  grid-template-columns: 64px 2fr 2fr 110px;
  font-size: 12.5px;
}
.table-header {
  background: var(--surface-2);
  border-bottom: 1px solid var(--border);
  color: var(--muted);
  font-size: 11.5px;
}
.table-header > div,
.table-row > div {
  padding: 8px 10px;
}
.table-row {
  border-bottom: 1px solid var(--border);
}
.table-row:last-child {
  border-bottom: none;
}
.table-row:hover {
  background: var(--surface-2);
}
.align-right {
  text-align: right;
}
.muted {
  color: var(--muted);
}
.ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.status-badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 11px;
  font-size: 11.5px;
}
.progress-row {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 12px;
  color: var(--muted);
}
.progress-track {
  flex: 1;
  height: 6px;
  border-radius: 3px;
  background: var(--surface-2);
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: var(--accent);
  transition: width 0.2s ease;
}
</style>
