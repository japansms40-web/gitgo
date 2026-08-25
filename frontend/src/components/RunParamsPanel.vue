<script setup>
import NumberStepper from './NumberStepper.vue'

const props = defineProps({
  intervalSec: { type: Number, required: true },
  retries: { type: Number, required: true },
  autoCreate: { type: Boolean, required: true },
  queueCount: { type: Number, required: true },
  running: { type: Boolean, required: true },
})
const emit = defineEmits([
  'update:intervalSec',
  'update:retries',
  'update:autoCreate',
  'start',
  'stop',
  'save-config',
  'clear-queue',
  'export-links',
])
</script>

<template>
  <div class="panel">
    <div class="panel-title">运行参数 RUN PARAMS</div>

    <NumberStepper
      label="发布间隔"
      unit="秒"
      :model-value="intervalSec"
      @update:model-value="emit('update:intervalSec', $event)"
    />
    <NumberStepper
      label="失败重试"
      unit="次"
      :model-value="retries"
      @update:model-value="emit('update:retries', $event)"
    />
    <NumberStepper label="队列篇数" unit="篇" :model-value="queueCount" :editable="false" />

    <label class="autocreate">
      <input type="checkbox" :checked="autoCreate" @change="emit('update:autoCreate', $event.target.checked)" />
      仓库不存在时自动创建
    </label>

    <div class="divider" />

    <button v-if="!running" class="btn-primary btn-tall" @click="emit('start')">开始发布</button>
    <template v-else>
      <button class="btn-danger-outline btn-tall" @click="emit('stop')">停止</button>
    </template>

    <div class="divider" />

    <div class="btn-grid">
      <button class="btn-outline" @click="emit('save-config')">保存配置</button>
      <button class="btn-outline" @click="emit('clear-queue')">清空队列</button>
      <button class="btn-outline btn-span2" @click="emit('export-links')">导出链接列表</button>
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
.autocreate {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  color: var(--text);
  cursor: pointer;
}
.divider {
  height: 1px;
  background: var(--border);
  margin: 2px 0;
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
.btn-danger-outline {
  border: 1px solid var(--err);
  color: var(--err);
  background: var(--surface);
}
.btn-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}
.btn-span2 {
  grid-column: span 2;
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
</style>
