<script setup>
defineProps({
  page: { type: String, required: true },
})
const emit = defineEmits(['navigate'])

const items = [
  { key: 'publish', cn: '发布', en: 'PUBLISH' },
  { key: 'content', cn: '内容设置', en: 'CONTENT' },
  { key: 'proxy', cn: 'IP 设置', en: 'PROXY' },
  { key: 'captcha', cn: '打码设置', en: 'CAPTCHA' },
  { key: 'spider', cn: '蜘蛛设置', en: 'SPIDER' },
  { key: 'links', cn: '其他设置', en: 'URL / LINKS' },
  { key: 'collect', cn: '采集文章', en: 'COLLECT' },
]
</script>

<template>
  <div class="nav">
    <div class="nav-title">工作台 WORKSPACE</div>

    <button v-for="item in items" :key="item.key" class="nav-item" :class="{ active: page === item.key }" @click="emit('navigate', item.key)">
      <svg v-if="item.key === 'publish'" width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <path d="M8 11V3.5M8 3.5L5.2 6.3M8 3.5L10.8 6.3M2.5 10.5v2a1 1 0 001 1h9a1 1 0 001-1v-2" />
      </svg>
      <svg v-else-if="item.key === 'content'" width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <rect x="3.2" y="2.5" width="9.6" height="11" rx="1.2" />
        <path d="M5.5 6h5M5.5 8.3h5M5.5 10.6h3" />
      </svg>
      <svg v-else-if="item.key === 'proxy'" width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <circle cx="8" cy="8" r="5.5" />
        <path d="M2.5 8h11M8 2.5c1.8 1.6 1.8 9.4 0 11M8 2.5c-1.8 1.6-1.8 9.4 0 11" />
      </svg>
      <svg v-else-if="item.key === 'captcha'" width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <path d="M8 2.3l4.7 1.7v3.6c0 3.2-2 5.7-4.7 6.6-2.7-.9-4.7-3.4-4.7-6.6V4L8 2.3z" stroke-linejoin="round" />
        <path d="M6 8.1l1.4 1.4L10.2 6.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
      <svg v-else-if="item.key === 'spider'" width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round">
        <path d="M8 2.5v2.4M8 11.1v2.4M2.5 8h2.4M11.1 8h2.4M4.3 4.3l1.7 1.7M10 10l1.7 1.7M11.7 4.3L10 6M6 10l-1.7 1.7" />
      </svg>
      <svg v-else-if="item.key === 'links'" width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round">
        <path d="M3 5h10M3 8h7M3 11h4" />
      </svg>
      <svg v-else-if="item.key === 'collect'" width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <path d="M8 3.5V11M8 11L5.2 8.2M8 11l2.8-2.8M2.5 12.5v1a1 1 0 001 1h9a1 1 0 001-1v-1" />
      </svg>
      <span class="nav-label">
        <span class="nav-label-cn">{{ item.cn }}</span>
        <span class="nav-label-en">{{ item.en }}</span>
      </span>
    </button>

    <div class="nav-spacer" />

    <button class="nav-item" :class="{ active: page === 'help' }" @click="emit('navigate', 'help')">
      <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <circle cx="8" cy="8" r="5.5" />
        <path d="M6.5 6.3c0-.9.7-1.5 1.5-1.5s1.5.6 1.5 1.5c0 .8-.7 1.1-1.1 1.4-.3.2-.4.5-.4.9" />
        <circle cx="8" cy="11" r=".7" fill="currentColor" stroke="none" />
      </svg>
      <span class="nav-label">
        <span class="nav-label-cn">使用说明</span>
        <span class="nav-label-en">HELP</span>
      </span>
    </button>
  </div>
</template>

<style scoped>
.nav {
  width: 182px;
  flex: 0 0 182px;
  background: var(--nav);
  border-right: 1px solid var(--border);
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
}
.nav-title {
  font-size: 10px;
  letter-spacing: 0.1em;
  color: var(--muted);
  padding: 2px 8px 6px;
  text-transform: uppercase;
}
.nav-spacer {
  flex: 1;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 40px;
  flex: 0 0 40px;
  padding: 0 10px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font-family: inherit;
  text-align: left;
  border-left: 3px solid transparent;
}
.nav-item svg {
  flex: 0 0 15px;
}
.nav-item:hover {
  background: var(--surface);
}
.nav-item.active {
  background: var(--accent-weak);
  color: var(--accent);
  border-left-color: var(--accent);
}
.nav-label {
  display: flex;
  flex-direction: column;
  line-height: 1.1;
  overflow: hidden;
}
.nav-label-cn {
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}
.nav-label-en {
  font-size: 9px;
  letter-spacing: 0.05em;
  color: var(--muted);
  white-space: nowrap;
}
.nav-item.active .nav-label-en {
  color: var(--accent);
  opacity: 0.75;
}
</style>
