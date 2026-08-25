<script setup>
defineProps({
  page: { type: String, required: true }, // 'publish' | 'help'
})
const emit = defineEmits(['navigate'])
</script>

<template>
  <div class="nav">
    <div class="nav-title">工作台 WORKSPACE</div>
    <button class="nav-item" :class="{ active: page === 'publish' }" @click="emit('navigate', 'publish')">
      <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <path d="M8 11V3.5M8 3.5L5.2 6.3M8 3.5L10.8 6.3M2.5 10.5v2a1 1 0 001 1h9a1 1 0 001-1v-2" />
      </svg>
      <span class="nav-label">
        <span class="nav-label-cn">发布</span>
        <span class="nav-label-en">PUBLISH</span>
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
}
.nav-label-cn {
  font-size: 13px;
  font-weight: 600;
}
.nav-label-en {
  font-size: 9px;
  letter-spacing: 0.05em;
  color: var(--muted);
}
.nav-item.active .nav-label-en {
  color: var(--accent);
  opacity: 0.75;
}
</style>
