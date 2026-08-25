<script setup>
import { NumberStepper } from '@dongfang/df-ui-shell'

defineProps({
  threads: { type: Number, required: true },
  interval: { type: Number, required: true },
  perAccount: { type: Number, required: true },
  failSwitch: { type: Number, required: true },
  accountCycles: { type: Number, required: true },
  roundInterval: { type: Number, required: true },
  keywordSlots: { type: Number, required: true },
  newRepo: { type: Boolean, required: true },
  working: { type: Boolean, required: true },
})
const emit = defineEmits([
  'update:threads',
  'update:interval',
  'update:perAccount',
  'update:failSwitch',
  'update:accountCycles',
  'update:roundInterval',
  'update:keywordSlots',
  'update:newRepo',
  'keyword-settings',
  'save-config',
  'clear-accounts',
  'account-feature',
  'view-links',
  'start-work',
])
</script>

<template>
  <div class="panel">
    <div class="panel-title">任务参数 RUN PARAMS</div>

    <NumberStepper
      label="线程数量"
      unit="个"
      :min="1"
      :model-value="threads"
      @update:model-value="emit('update:threads', $event)"
    />
    <NumberStepper
      label="发布间隔"
      unit="秒"
      :min="0"
      :model-value="interval"
      @update:model-value="emit('update:interval', $event)"
    />
    <NumberStepper
      label="每号发布"
      unit="次"
      :min="1"
      :model-value="perAccount"
      @update:model-value="emit('update:perAccount', $event)"
    />
    <NumberStepper
      label="失败换号"
      unit="次"
      :min="1"
      :model-value="failSwitch"
      @update:model-value="emit('update:failSwitch', $event)"
    />
    <NumberStepper
      label="账号循环"
      unit="轮"
      :min="1"
      :model-value="accountCycles"
      @update:model-value="emit('update:accountCycles', $event)"
    />
    <NumberStepper
      label="每轮间隔"
      unit="秒"
      :min="0"
      :model-value="roundInterval"
      @update:model-value="emit('update:roundInterval', $event)"
    />
    <NumberStepper
      label="关键词位"
      unit="个"
      :min="1"
      :model-value="keywordSlots"
      @update:model-value="emit('update:keywordSlots', $event)"
    />

    <label class="checkbox-row">
      <input type="checkbox" class="sr-only" :checked="newRepo" @change="emit('update:newRepo', $event.target.checked)" />
      <span class="checkbox-box" :class="{ checked: newRepo }">
        <svg v-if="newRepo" width="10" height="10" viewBox="0 0 16 16" fill="none" stroke="#fff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 8.2l3.3 3.3L13 4.5" />
        </svg>
      </span>
      创建仓库 New repo
    </label>

    <div class="divider" />

    <button class="btn-primary btn-full" @click="emit('keyword-settings')">关键词设置</button>

    <div class="btn-grid">
      <button class="btn-outline btn-full" @click="emit('save-config')">保存配置</button>
      <button class="btn-outline btn-full" @click="emit('clear-accounts')">清空账号</button>
      <button class="btn-outline btn-full" @click="emit('account-feature')">换号特征</button>
      <button class="btn-outline btn-full" @click="emit('view-links')">查看链接</button>
    </div>

    <div class="spacer" />

    <button class="btn-primary btn-tall" :disabled="working" @click="emit('start-work')">
      {{ working ? '工作中…' : '开始工作' }}
    </button>
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
  min-height: 8px;
}
.btn-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
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
  border: none;
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
.btn-primary:hover:not(:disabled) {
  opacity: 0.92;
}
button:disabled {
  opacity: 0.5;
  cursor: default;
}
</style>
