<script setup>
defineProps({
  url: { type: String, required: true },
  enabled: { type: Boolean, required: true },
  testing: { type: Boolean, default: false },
  // 拨测结果：null 或 { ok, message, statusCode, latencyMs }
  testResult: { type: Object, default: null },
})
const emit = defineEmits(['update:url', 'update:enabled', 'test', 'save'])
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">
        <span class="page-title-main">代理 Proxy</span>
        <span class="page-title-sub">全局 socks5 / http 代理，启用后所有 GitHub 请求经此转发</span>
      </div>
      <div class="spacer" />
      <button class="btn-primary" @click="emit('save')">保存</button>
    </div>

    <div class="body">
      <div class="card">
        <label class="field">
          <span class="field-label">代理地址</span>
          <input
            class="input"
            :value="url"
            spellcheck="false"
            autocomplete="off"
            placeholder="socks5://127.0.0.1:1080   或   http://user:pass@host:port"
            @input="emit('update:url', $event.target.value)"
          />
        </label>
        <p class="hint">支持 <code>socks5://</code> / <code>http://</code> / <code>https://</code>，可带 <code>user:pass@</code>。留空 = 直连。</p>

        <label class="checkbox-row">
          <input type="checkbox" class="sr-only" :checked="enabled" @change="emit('update:enabled', $event.target.checked)" />
          <span class="checkbox-box" :class="{ checked: enabled }">
            <svg v-if="enabled" width="10" height="10" viewBox="0 0 16 16" fill="none" stroke="#fff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 8.2l3.3 3.3L13 4.5" />
            </svg>
          </span>
          启用代理
        </label>

        <div class="divider" />

        <div class="actions">
          <button class="btn-outline" :disabled="testing" @click="emit('test')">
            {{ testing ? '测试中…' : '测试连通' }}
          </button>
          <span
            v-if="testResult"
            class="result"
            :class="testResult.ok ? 'ok' : 'err'"
          >
            <span class="dot" />
            {{ testResult.message }}
          </span>
          <span v-else class="result muted">通过代理拨测 github.com，验证是否可达</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: var(--bg);
}
.page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border);
}
.page-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.page-title-main {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
}
.page-title-sub {
  font-size: 11.5px;
  color: var(--muted);
}
.spacer {
  flex: 1;
}
.btn-primary {
  height: 32px;
  padding: 0 16px;
  border-radius: 6px;
  border: none;
  background: var(--accent);
  color: #fff;
  font-size: 13px;
  cursor: pointer;
}
.btn-primary:hover {
  opacity: 0.92;
}
.body {
  flex: 1;
  overflow: auto;
  padding: 18px;
}
.card {
  max-width: 640px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.field-label {
  font-size: 12.5px;
  color: var(--text);
  font-weight: 500;
}
.input {
  height: 36px;
  padding: 0 12px;
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  background: var(--surface-2);
  color: var(--text);
  font-size: 13px;
  outline: none;
}
.input:focus {
  border-color: var(--accent);
}
.hint {
  margin: 0;
  font-size: 11.5px;
  color: var(--muted);
  line-height: 1.6;
}
.hint code {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  padding: 1px 4px;
  border-radius: 3px;
  background: var(--surface-2);
  color: var(--text);
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
.divider {
  height: 1px;
  background: var(--border);
  margin: 2px 0;
}
.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.btn-outline {
  height: 34px;
  padding: 0 16px;
  border-radius: 6px;
  border: 1px solid var(--border-strong);
  background: var(--surface);
  color: var(--text);
  font-size: 12.5px;
  cursor: pointer;
}
.btn-outline:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
}
.btn-outline:disabled {
  opacity: 0.5;
  cursor: default;
}
.result {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12.5px;
}
.result.ok {
  color: var(--ok);
}
.result.err {
  color: var(--err);
}
.result.muted {
  color: var(--muted);
}
.dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}
</style>
