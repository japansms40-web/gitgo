<script setup>
import NumberStepper from './NumberStepper.vue'

const props = defineProps({
  threads: { type: Number, required: true },
  intervalSec: { type: Number, required: true },
  perAccountCount: { type: Number, required: true },
  failSwitchCount: { type: Number, required: true },
  cycleRounds: { type: Number, required: true },
  roundIntervalSec: { type: Number, required: true },
  keywordSlots: { type: Number, required: true },
  createRepo: { type: Boolean, required: true },
  contentCount: { type: Number, required: true },
  running: { type: Boolean, required: true },
  paused: { type: Boolean, required: true },
  round: { type: Number, required: true },
  roundDone: { type: Number, required: true },
  roundTotal: { type: Number, required: true },
})
const emit = defineEmits([
  'update:threads',
  'update:intervalSec',
  'update:perAccountCount',
  'update:failSwitchCount',
  'update:cycleRounds',
  'update:roundIntervalSec',
  'update:keywordSlots',
  'update:createRepo',
  'select-folder',
  'select-files',
  'clear-content',
  'start',
  'pause',
  'resume',
  'stop',
  'save-config',
  'clear-accounts',
  'keyword-settings',
  'switch-profile',
  'view-results',
])

const progressPercent = (done, total) => (total === 0 ? 0 : Math.round((done / total) * 100))
</script>

<template>
  <div class="panel">
    <div class="content-box">
      <div class="content-row">
        <span class="content-label">发布内容</span>
        <span class="content-count mono">{{ contentCount }} 篇</span>
      </div>
      <div class="content-actions">
        <button class="btn-outline" @click="emit('select-folder')">选择文件夹</button>
        <button class="btn-outline" @click="emit('select-files')">添加文件</button>
      </div>
      <button v-if="contentCount > 0" class="link-danger" @click="emit('clear-content')">清空内容</button>
    </div>

    <div class="divider" />

    <div class="panel-title">任务参数 RUN PARAMS</div>

    <NumberStepper label="线程数量" unit="个" :model-value="threads" @update:model-value="emit('update:threads', $event)" />
    <NumberStepper label="发布间隔" unit="秒" :model-value="intervalSec" @update:model-value="emit('update:intervalSec', $event)" />
    <NumberStepper label="每号发布" unit="次" :model-value="perAccountCount" @update:model-value="emit('update:perAccountCount', $event)" />
    <NumberStepper label="失败换号" unit="次" :model-value="failSwitchCount" @update:model-value="emit('update:failSwitchCount', $event)" />
    <NumberStepper label="账号循环" unit="轮" :model-value="cycleRounds" @update:model-value="emit('update:cycleRounds', $event)" />
    <NumberStepper label="每轮间隔" unit="秒" :model-value="roundIntervalSec" @update:model-value="emit('update:roundIntervalSec', $event)" />
    <NumberStepper label="关键词位" unit="个" :model-value="keywordSlots" @update:model-value="emit('update:keywordSlots', $event)" />

    <label class="checkbox-row">
      <input type="checkbox" class="sr-only" :checked="createRepo" @change="emit('update:createRepo', $event.target.checked)" />
      <span class="checkbox-box" :class="{ checked: createRepo }">
        <svg v-if="createRepo" width="10" height="10" viewBox="0 0 16 16" fill="none" stroke="#fff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 8.2l3.3 3.3L13 4.5" />
        </svg>
      </span>
      创建仓库 New repo
    </label>

    <div class="divider" />

    <button class="btn-primary btn-full" @click="emit('keyword-settings')">关键词设置</button>

    <div class="btn-grid">
      <button class="btn-outline" @click="emit('save-config')">保存配置</button>
      <button class="btn-outline" @click="emit('clear-accounts')">清空账号</button>
      <button class="btn-outline" @click="emit('switch-profile')">换号特征</button>
      <button class="btn-outline" @click="emit('view-results')">查看链接</button>
    </div>

    <div class="spacer" />

    <div v-if="running" class="round-box">
      <div class="round-row">
        <span>本轮进度</span>
        <span class="mono">{{ roundDone }} / {{ roundTotal }}</span>
      </div>
      <div class="progress-track">
        <div class="progress-fill" :style="{ width: progressPercent(roundDone, roundTotal) + '%' }" />
      </div>
    </div>

    <button v-if="!running" class="btn-primary btn-tall" @click="emit('start')">开始发布</button>
    <div v-else class="run-actions">
      <button v-if="!paused" class="btn-pause" @click="emit('pause')">
        <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor"><rect x="4" y="3" width="2.6" height="10" rx="0.5" /><rect x="9.4" y="3" width="2.6" height="10" rx="0.5" /></svg>
        暂停
      </button>
      <button v-else class="btn-resume" @click="emit('resume')">
        <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor"><path d="M4.5 3v10l9-5-9-5z" /></svg>
        继续
      </button>
      <button class="btn-danger-outline" @click="emit('stop')">停止</button>
    </div>
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
.content-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}
.link-danger {
  align-self: flex-start;
  border: none;
  background: none;
  padding: 0;
  color: var(--err);
  font-size: 12px;
  cursor: pointer;
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
.btn-full {
  width: 100%;
  height: 34px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.btn-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}
.btn-outline {
  height: 30px;
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
.round-box {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.round-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  color: var(--muted);
}
.progress-track {
  height: 6px;
  border-radius: 3px;
  background: var(--surface);
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: var(--accent);
  transition: width 0.2s ease;
}
.run-actions {
  display: flex;
  gap: 8px;
}
.btn-pause,
.btn-resume {
  flex: 1;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  border-radius: 6px;
  border: none;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  color: #fff;
}
.btn-pause {
  background: var(--warn);
}
.btn-resume {
  background: var(--ok);
}
.btn-danger-outline {
  flex: 1;
  height: 44px;
  border-radius: 6px;
  border: 1px solid var(--err);
  color: var(--err);
  background: var(--surface);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
</style>
