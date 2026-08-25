<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime'

const props = defineProps({
  accounts: { type: Array, required: true }, // {ck, ua, ip, status, success, fail, total, bad}
})
const emit = defineEmits([
  'import-account-click',
  'import-account-files',
  'paste-clipboard',
  'remove-account',
  'mark-bad',
  'clear-accounts',
  'export-result',
  'test-account',
  'copy-ck',
  'save-config',
])

const STATUS_STYLE = {
  待发: { bg: 'transparent', fg: 'var(--muted)' },
  发布中: { bg: 'rgba(154,103,0,.16)', fg: 'var(--warn)' },
  成功: { bg: 'rgba(26,127,55,.14)', fg: 'var(--ok)' },
  失败: { bg: 'rgba(207,34,46,.12)', fg: 'var(--err)' },
}
function statusStyle(status) {
  return STATUS_STYLE[status] || STATUS_STYLE['待发']
}

function truncate(text, max = 18) {
  if (!text) return '—'
  return text.length > max ? text.slice(0, max) + '…' : text
}

const search = ref('')
const activeTab = ref('all')

const rows = computed(() => props.accounts.map((a, index) => ({ ...a, index })))

const counts = computed(() => {
  let success = 0
  let fail = 0
  for (const a of props.accounts) {
    if (a.status === '成功') success++
    else if (a.status === '失败') fail++
  }
  return {
    all: props.accounts.length,
    success,
    fail,
    pending: props.accounts.length - success - fail,
  }
})

const tabs = computed(() => [
  { key: 'all', label: '全部', count: counts.value.all },
  { key: 'success', label: '成功', count: counts.value.success },
  { key: 'fail', label: '失败', count: counts.value.fail },
  { key: 'pending', label: '待发', count: counts.value.pending },
])

const filteredRows = computed(() => {
  let list = rows.value
  if (activeTab.value === 'success') list = list.filter((a) => a.status === '成功')
  else if (activeTab.value === 'fail') list = list.filter((a) => a.status === '失败')
  else if (activeTab.value === 'pending') list = list.filter((a) => a.status !== '成功' && a.status !== '失败')

  const q = search.value.trim().toLowerCase()
  if (!q) return list
  return list.filter(
    (a) => a.ck.toLowerCase().includes(q) || (a.ip || '').toLowerCase().includes(q) || a.status.toLowerCase().includes(q)
  )
})

// ---- 右键菜单 ----
const menu = reactive({ show: false, x: 0, y: 0, index: -1 })
function openMenu(evt, index) {
  menu.show = true
  menu.x = evt.clientX
  menu.y = evt.clientY
  menu.index = index
}
function closeMenu() {
  menu.show = false
}
function menuAction(action) {
  const i = menu.index
  closeMenu()
  if (i < 0) return
  if (action === 'copy') emit('copy-ck', props.accounts[i]?.ck)
  else if (action === 'test') emit('test-account', i)
  else if (action === 'remove') emit('remove-account', i)
  else if (action === 'bad') emit('mark-bad', i)
}

// ---- 拖入 TXT 批量导入 ----
onMounted(() => {
  OnFileDrop((x, y, paths) => {
    const txtPaths = paths.filter((p) => p.toLowerCase().endsWith('.txt'))
    if (txtPaths.length) emit('import-account-files', txtPaths)
  }, true)
  window.addEventListener('click', closeMenu)
})
onUnmounted(() => {
  OnFileDropOff()
  window.removeEventListener('click', closeMenu)
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">
        <span class="page-title-main">发布 Publish</span>
        <span class="page-title-sub">账号队列与发布结果，双击粘贴剪贴板 · 拖入 TXT 批量导入</span>
      </div>
      <div class="spacer" />
      <div class="page-actions">
        <button class="btn-outline" @click="emit('import-account-click')">导入账号</button>
        <button class="btn-outline" @click="emit('export-result')">导出结果</button>
        <button class="btn-primary" @click="emit('save-config')">保存配置</button>
      </div>
    </div>

    <div class="toolbar">
      <div class="search-box">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
          <circle cx="7" cy="7" r="4.6" /><path d="M10.4 10.4L14 14" stroke-linecap="round" />
        </svg>
        <input v-model="search" placeholder="搜索 CK / IP / 状态…" />
      </div>
      <div class="tabs">
        <button
          v-for="t in tabs"
          :key="t.key"
          class="tab"
          :class="{ active: activeTab === t.key }"
          @click="activeTab = t.key"
        >{{ t.label }} {{ t.count }}</button>
      </div>
      <div class="spacer" />
      <button class="btn-outline" @dblclick="emit('paste-clipboard')">双击粘贴剪贴板</button>
      <button class="btn-outline btn-danger-text" @click="emit('clear-accounts')">清空账号</button>
    </div>

    <div class="table-wrap">
      <table v-if="accounts.length">
        <thead>
          <tr>
            <th class="col-index">序号</th>
            <th>CK</th>
            <th>UA</th>
            <th>IP</th>
            <th class="align-right">状态</th>
            <th class="align-right">成功</th>
            <th class="align-right">失败</th>
            <th class="align-right">总数</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in filteredRows" :key="row.index" @contextmenu.prevent="openMenu($event, row.index)">
            <td class="col-index mono muted">{{ row.index + 1 }}</td>
            <td class="mono ellipsis" :title="row.ck">{{ truncate(row.ck) }}</td>
            <td class="ellipsis muted">{{ row.ua || '—' }}</td>
            <td class="mono muted">{{ row.ip || '—' }}</td>
            <td class="align-right">
              <span class="status-badge" :style="{ background: statusStyle(row.status).bg, color: statusStyle(row.status).fg }">{{ row.status }}</span>
            </td>
            <td class="align-right mono">{{ row.total > 0 ? row.success : '—' }}</td>
            <td class="align-right mono">{{ row.total > 0 ? row.fail : '—' }}</td>
            <td class="align-right mono">{{ row.total > 0 ? row.total : '—' }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">还没有账号，拖入 TXT 文件、点「导入账号」，或双击「双击粘贴剪贴板」导入</div>
    </div>

    <div class="hint-row">右键行：复制 CK · 单独测试 · 移出列表 · 标记为坏号 | 拖入文本文件可批量导入，分隔符 ----</div>

    <div v-if="menu.show" class="context-menu" :style="{ left: menu.x + 'px', top: menu.y + 'px' }" @click.stop>
      <div class="menu-item" @click="menuAction('copy')">复制 CK</div>
      <div class="menu-item" @click="menuAction('test')">单独测试</div>
      <div class="menu-item" @click="menuAction('remove')">移出列表</div>
      <div class="menu-item menu-item-danger" @click="menuAction('bad')">标记为坏号</div>
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
  --wails-drop-target: drop;
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
  white-space: nowrap;
}
.btn-outline:hover {
  border-color: var(--accent);
  color: var(--accent);
}
.btn-danger-text {
  color: var(--err);
  border-color: var(--border);
}
.btn-danger-text:hover {
  border-color: var(--err);
  color: var(--err);
}
.btn-primary {
  height: 30px;
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
.toolbar {
  flex: 0 0 auto;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 10px;
  padding: 12px 18px;
  border-bottom: 1px solid var(--border);
}
.search-box {
  flex: 0 0 200px;
  display: flex;
  align-items: center;
  gap: 7px;
  height: 32px;
  padding: 0 10px;
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  background: var(--surface);
  color: var(--muted);
}
.search-box input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 12.5px;
  color: var(--text);
  font-family: inherit;
}
.tabs {
  display: flex;
  gap: 6px;
  flex: 0 0 auto;
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
  flex: 0 0 auto;
}
.tab.active {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-weak);
}
.table-wrap {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12.5px;
}
thead th {
  position: sticky;
  top: 0;
  background: var(--surface-2);
  border-bottom: 1px solid var(--border);
  color: var(--muted);
  font-size: 11.5px;
  font-weight: 500;
  text-align: left;
  padding: 9px 14px;
}
tbody td {
  padding: 9px 14px;
  border-bottom: 1px solid var(--border);
}
tbody tr:hover {
  background: var(--surface-2);
}
.col-index {
  width: 56px;
}
.align-right {
  text-align: right;
}
.muted {
  color: var(--muted);
}
.ellipsis {
  max-width: 220px;
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
.empty {
  padding: 40px 18px;
  text-align: center;
  color: var(--muted);
  font-size: 13px;
}
.hint-row {
  flex: 0 0 auto;
  padding: 9px 18px;
  border-top: 1px solid var(--border);
  color: var(--muted);
  font-size: 11.5px;
}
.context-menu {
  position: fixed;
  z-index: 100;
  min-width: 128px;
  padding: 4px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 7px;
  box-shadow: var(--shadow);
}
.menu-item {
  padding: 7px 10px;
  border-radius: 5px;
  font-size: 12.5px;
  cursor: pointer;
}
.menu-item:hover {
  background: var(--surface-2);
}
.menu-item-danger {
  color: var(--err);
}
</style>
