<script setup>
import { NumberStepper } from '@dongfang/df-ui-shell'

defineProps({
  count: { type: Number, required: true },
  dedupeLines: { type: Boolean, required: true },
  chineseOnly: { type: Boolean, required: true },
  draftCount: { type: Number, required: true },
  generating: { type: Boolean, required: true },
})
const emit = defineEmits([
  'update:count',
  'update:dedupeLines',
  'update:chineseOnly',
  'generate',
  'export',
  'save-config',
  'open-dir',
])
</script>

<template>
  <div class="panel">
    <div class="content-box">
      <div class="content-row">
        <span>生成结果</span>
        <span class="content-count mono">{{ draftCount }} 篇</span>
      </div>
      <button class="btn-outline btn-full" @click="emit('open-dir')">打开素材目录</button>
    </div>

    <div class="divider" />

    <div class="panel-title">生成参数 GENERATE</div>

    <NumberStepper
      label="生成篇数"
      unit="篇"
      :min="1"
      :model-value="count"
      @update:model-value="emit('update:count', $event)"
    />

    <label class="checkbox-row">
      <input type="checkbox" class="sr-only" :checked="dedupeLines" @change="emit('update:dedupeLines', $event.target.checked)" />
      <span class="checkbox-box" :class="{ checked: dedupeLines }">
        <svg v-if="dedupeLines" width="10" height="10" viewBox="0 0 16 16" fill="none" stroke="#fff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 8.2l3.3 3.3L13 4.5" />
        </svg>
      </span>
      去除重复行
    </label>

    <label class="checkbox-row">
      <input type="checkbox" class="sr-only" :checked="chineseOnly" @change="emit('update:chineseOnly', $event.target.checked)" />
      <span class="checkbox-box" :class="{ checked: chineseOnly }">
        <svg v-if="chineseOnly" width="10" height="10" viewBox="0 0 16 16" fill="none" stroke="#fff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 8.2l3.3 3.3L13 4.5" />
        </svg>
      </span>
      仅保留中文行
    </label>

    <div class="divider" />

    <button class="btn-outline btn-full" @click="emit('save-config')">保存配置</button>

    <div class="spacer" />

    <button class="btn-primary btn-tall" :disabled="generating" @click="emit('generate')">
      {{ generating ? '生成中…' : '生成' }}
    </button>
    <button class="btn-outline btn-full" :disabled="draftCount === 0" @click="emit('export')">导出 .md</button>
  </div>
</template>

<style scoped>
.panel {
  width: 264px;
  flex: 0 0 264px;
  background: var(--surface-2);
  border-left: 1px solid var(--border);
  padding: 13px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow: auto;
}
.panel-title {
  font-size: 10px;
  letter-spacing: 0.1em;
  color: var(--muted);
  text-transform: uppercase;
}
.content-box {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
}
.content-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12.5px;
}
.content-count {
  color: var(--accent);
  font-weight: 600;
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
  flex: 0 0 auto;
}
.spacer {
  flex: 1;
}
.btn-full {
  width: 100%;
  height: 34px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}
.btn-tall {
  height: 44px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  border: none;
}
.btn-primary {
  background: var(--accent);
  color: #fff;
}
.btn-outline {
  border: 1px solid var(--border-strong);
  background: var(--surface);
  font-size: 12.5px;
  color: var(--text);
}
.btn-outline:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
}
button:disabled {
  opacity: 0.5;
  cursor: default;
}
</style>
