<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'

const props = defineProps({
  accounts: { type: Array, required: true }, // { id, ck, ua, ip, status, success, fail, total, bad }
})
const emit = defineEmits([
  'import-file', // 点「导入账号」：让父组件弹文件框
  'paste-clipboard', // 双击 / 点按钮：从剪贴板导入
  'import-raw', // 拖入 TXT：把读到的原始文本交给父组件解析
  'export-results',
  'save-config',
  'clear-accounts',
  'copy-ck',
  'test-account',
  'remove-account',
  'mark-bad',
])

// 状态元数据：中文标签 + 样式类；发布模拟只会落在这五种之一。
const STATUS_META = {
  pending: { label: '待发', cls: 'muted' },
  publishing: { label: '发布中', cls: 'warn' },
  success: { label: '成功', cls: 'ok' },
  failed: { label: '失败', cls: 'err' },
  bad: { label: '坏号', cls: 'bad' },
}
function statusLabel(s) {
  return STATUS_META[s]?.label ?? s
}

// 顶部筛选胶囊：key 为 null 表示「全部」。
const filter = ref(null)
const search = ref('')

const counts = computed(() => {
  const c = { all: props.accounts.length, success: 0, failed: 0, pending: 0 }
  for (const a of props.accounts) {
    if (a.status === 'success') c.success++
    else if (a.status === 'failed') c.failed++
    else if (a.status === 'pending') c.pending++
  }
  return c
})

const chips = computed(() => [
  { key: null, label: '全部', n: counts.value.all, cls: 'accent' },
  { key: 'success', label: '成功', n: counts.value.success, cls: 'ok' },
  { key: 'failed', label: '失败', n: counts.value.failed, cls: 'err' },
  { key: 'pending', label: '待发', n: counts.value.pending, cls: 'muted' },
])

// 展示用列表：先按筛选，再按搜索词匹配 CK / IP / UA / 状态。
const rows = computed(() => {
  const q = search.value.trim().toLowerCase()
  return props.accounts
    .map((a, i) => ({ a, seq: i + 1 }))
    .filter(({ a }) => (filter.value ? a.status === filter.value : true))
    .filter(({ a }) => {
      if (!q) return true
      return (
        a.ck.toLowerCase().includes(q) ||
        (a.ip || '').toLowerCase().includes(q) ||
        (a.ua || '').toLowerCase().includes(q) ||
        statusLabel(a.status).includes(q)
      )
    })
})

// 只有动过（发过或非待发）的账号才展示成功/失败/总数，否则显示「—」，对齐截图。
function hasStats(a) {
  return a.total > 0 || (a.status !== 'pending' && a.status !== 'bad')
}
function dash(v) {
  return v ? v : '—'
}
function shortCk(ck) {
  return ck.length > 16 ? ck.slice(0, 16) + '…' : ck
}

// ---------- 右键菜单 ----------
const menu = reactive({ visible: false, x: 0, y: 0, account: null })
function openMenu(e, a) {
  menu.visible = true
  menu.x = e.clientX
  menu.y = e.clientY
  menu.account = a
}
function closeMenu() {
  menu.visible = false
  menu.account = null
}
function runMenu(action) {
  const a = menu.account
  closeMenu()
  if (!a) return
  emit(action, a)
}
function onKey(e) {
  if (e.key === 'Escape') closeMenu()
}
onMounted(() => {
  window.addEventListener('click', closeMenu)
  window.addEventListener('keydown', onKey)
})
onBeforeUnmount(() => {
  window.removeEventListener('click', closeMenu)
  window.removeEventListener('keydown', onKey)
})

// ---------- 拖入 TXT ----------
const dragging = ref(false)
function onDragOver() {
  dragging.value = true
}
function onDragLeave() {
  dragging.value = false
}
async function onDrop(e) {
  dragging.value = false
  const file = e.dataTransfer?.files?.[0]
  if (!file) return
  try {
    const text = await file.text()
    emit('import-raw', text)
  } catch {
    /* 读取失败就忽略，用户可改用「导入账号」按钮 */
  }
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">
        <span class="page-title-main">发布 Publish</span>
        <span class="page-title-sub">账号队列与发布结果，双击粘贴剪贴板 · 拖入 TXT 批量导入</span>
      </div>
      <div class="spacer" />
      <button class="btn-outline" @click="emit('import-file')">导入账号</button>
      <button class="btn-outline" @click="emit('export-results')">导出结果</button>
      <button class="btn-primary" @click="emit('save-config')">保存配置</button>
    </div>

    <div class="toolbar">
      <div class="search">
        <svg class="search-icon" width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
          <circle cx="7" cy="7" r="4.5" />
          <path d="M10.5 10.5L14 14" stroke-linecap="round" />
        </svg>
        <input v-model="search" class="search-input" placeholder="搜索 CK / IP / 状态…" />
      </div>

      <div class="chips">
        <button
          v-for="c in chips"
          :key="c.label"
          class="chip"
          :class="[{ active: filter === c.key }, c.cls]"
          @click="filter = c.key"
        >
          {{ c.label }} <span class="chip-n">{{ c.n }}</span>
        </button>
      </div>

      <div class="spacer" />
      <button class="btn-outline" @click="emit('paste-clipboard')">双击粘贴剪贴板</button>
      <button class="btn-outline danger" @click="emit('clear-accounts')">清空账号</button>
    </div>

    <div
      class="page-body"
      :class="{ dragging }"
      @dblclick="emit('paste-clipboard')"
      @dragover.prevent="onDragOver"
      @dragleave="onDragLeave"
      @drop.prevent="onDrop"
    >
      <div v-if="accounts.length === 0" class="empty">
        还没有账号，点右上「导入账号」、按钮「双击粘贴剪贴板」，或直接把 TXT 拖进来
      </div>

      <div v-else class="table-wrap">
        <table class="tbl">
          <thead>
            <tr>
              <th class="c-seq">序号</th>
              <th class="c-ck">CK</th>
              <th class="c-ua">UA</th>
              <th class="c-ip">IP</th>
              <th class="c-status">状态</th>
              <th class="c-num">成功</th>
              <th class="c-num">失败</th>
              <th class="c-num">总数</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="{ a, seq } in rows" :key="a.id" :class="{ bad: a.bad }" @contextmenu.prevent="openMenu($event, a)">
              <td class="c-seq muted">{{ seq }}</td>
              <td class="c-ck mono" :title="a.ck">{{ shortCk(a.ck) }}</td>
              <td class="c-ua mono muted">{{ dash(a.ua) }}</td>
              <td class="c-ip mono muted">{{ dash(a.ip) }}</td>
              <td class="c-status">
                <span class="pill" :class="STATUS_META[a.status]?.cls">{{ statusLabel(a.status) }}</span>
              </td>
              <template v-if="hasStats(a)">
                <td class="c-num num-ok">{{ a.success }}</td>
                <td class="c-num num-err">{{ a.fail }}</td>
                <td class="c-num">{{ a.total }}</td>
              </template>
              <template v-else>
                <td class="c-num muted">—</td>
                <td class="c-num muted">—</td>
                <td class="c-num muted">—</td>
              </template>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="foot">
      右键行：复制 CK · 单独测试 · 移出列表 · 标记为坏号 | 拖入文本文件可批量导入，每行一个 CK
    </div>

    <!-- 右键菜单 -->
    <div
      v-if="menu.visible"
      class="ctx"
      :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
      @click.stop
    >
      <button class="ctx-item" @click="runMenu('copy-ck')">复制 CK</button>
      <button class="ctx-item" @click="runMenu('test-account')">单独测试</button>
      <button class="ctx-item" @click="runMenu('remove-account')">移出列表</button>
      <button class="ctx-item danger" @click="runMenu('mark-bad')">标记为坏号</button>
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
  gap: 8px;
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

.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border-bottom: 1px solid var(--border);
}
.search {
  position: relative;
  flex: 0 0 300px;
}
.search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--muted);
  pointer-events: none;
}
.search-input {
  width: 100%;
  height: 34px;
  padding: 0 10px 0 30px;
  border: 1px solid var(--border-strong);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  font-family: inherit;
  font-size: 12.5px;
  outline: none;
}
.search-input:focus {
  border-color: var(--accent);
}

.chips {
  display: flex;
  gap: 6px;
}
.chip {
  height: 30px;
  padding: 0 12px;
  border-radius: 16px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  font-size: 12.5px;
  cursor: pointer;
  white-space: nowrap;
}
.chip .chip-n {
  color: var(--muted);
  font-weight: 600;
  margin-left: 2px;
}
.chip.active.accent {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-weak);
}
.chip.active.accent .chip-n {
  color: var(--accent);
}
.chip.active.ok {
  border-color: var(--ok);
  color: var(--ok);
}
.chip.active.ok .chip-n {
  color: var(--ok);
}
.chip.active.err {
  border-color: var(--err);
  color: var(--err);
}
.chip.active.err .chip-n {
  color: var(--err);
}
.chip.active.muted {
  border-color: var(--border-strong);
}

.btn-primary,
.btn-outline {
  height: 32px;
  padding: 0 14px;
  border-radius: 6px;
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  font-family: inherit;
}
.btn-primary {
  border: none;
  background: var(--accent);
  color: #fff;
}
.btn-primary:hover {
  opacity: 0.92;
}
.btn-outline {
  border: 1px solid var(--border-strong);
  background: var(--surface);
  color: var(--text);
}
.btn-outline:hover {
  border-color: var(--accent);
  color: var(--accent);
}
.btn-outline.danger {
  color: var(--err);
}
.btn-outline.danger:hover {
  border-color: var(--err);
  color: var(--err);
}

.page-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 12px 18px;
}
.page-body.dragging {
  outline: 2px dashed var(--accent);
  outline-offset: -8px;
  background: var(--accent-weak);
}
.empty {
  padding: 60px 0;
  text-align: center;
  color: var(--muted);
  font-size: 13px;
}

.table-wrap {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: auto;
}
.tbl {
  width: 100%;
  border-collapse: collapse;
  font-size: 12.5px;
}
.tbl thead th {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--surface-2);
  color: var(--muted);
  font-weight: 600;
  text-align: left;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
}
.tbl tbody td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
}
.tbl tbody tr:last-child td {
  border-bottom: none;
}
.tbl tbody tr:hover {
  background: var(--surface-2);
}
.tbl tbody tr.bad td {
  opacity: 0.5;
  text-decoration: line-through;
}
.muted {
  color: var(--muted);
}
.c-seq {
  width: 56px;
}
.c-ck {
  min-width: 160px;
}
.c-ua {
  width: 170px;
}
.c-ip {
  width: 140px;
}
.c-status {
  width: 92px;
}
.c-num {
  width: 66px;
  text-align: right;
}
.num-ok {
  color: var(--ok);
}
.num-err {
  color: var(--err);
}

.pill {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 11.5px;
  line-height: 1.3;
}
.pill.ok {
  background: color-mix(in srgb, var(--ok) 15%, transparent);
  color: var(--ok);
}
.pill.err {
  background: color-mix(in srgb, var(--err) 15%, transparent);
  color: var(--err);
}
.pill.warn {
  background: color-mix(in srgb, var(--warn) 18%, transparent);
  color: var(--warn);
}
.pill.muted {
  background: var(--surface-2);
  color: var(--muted);
  border: 1px solid var(--border);
}
.pill.bad {
  background: color-mix(in srgb, var(--err) 12%, transparent);
  color: var(--err);
}

.foot {
  flex: 0 0 auto;
  padding: 8px 18px 10px;
  border-top: 1px solid var(--border);
  color: var(--muted);
  font-size: 11.5px;
}

.ctx {
  position: fixed;
  z-index: 100;
  min-width: 140px;
  padding: 4px;
  background: var(--surface);
  border: 1px solid var(--border-strong);
  border-radius: 8px;
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
}
.ctx-item {
  height: 32px;
  padding: 0 10px;
  border: none;
  background: transparent;
  color: var(--text);
  font-size: 12.5px;
  text-align: left;
  border-radius: 5px;
  cursor: pointer;
  font-family: inherit;
}
.ctx-item:hover {
  background: var(--accent-weak);
  color: var(--accent);
}
.ctx-item.danger {
  color: var(--err);
}
.ctx-item.danger:hover {
  background: color-mix(in srgb, var(--err) 12%, transparent);
  color: var(--err);
}
</style>
