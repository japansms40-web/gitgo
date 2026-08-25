<script setup>
const items = [
  { label: '内容', body: '右侧「选择文件夹/添加文件」导入 .md / .txt 作为待发布内容，文件名作为标题。' },
  { label: '账号', body: '发布页里导入账号（CK/UA/IP）：点「导入账号」选 TXT 文件，或双击「双击粘贴剪贴板」粘贴，多个账号用 ---- 分隔。' },
  { label: '发布', body: '开始发布后，账号队列逐个发布一篇内容（内容不够会循环使用），标记为坏号的账号会被跳过。' },
  { label: '重试', body: '可设置账号间的发布间隔与单账号失败重试次数。' },
  { label: '统计', body: '每个账号的成功/失败/总数是长期累计值，账号队列本身也会保存在本机，重启不丢失。' },
  { label: '管理', body: '右键账号行可复制 CK、单独测试、移出列表或标记为坏号；顶部可搜索、按状态筛选、导出结果。' },
]
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">
        <span class="page-title-main">使用说明</span>
        <span class="page-title-sub">格式约定、常用操作与环境要求</span>
      </div>
    </div>
    <div class="page-body">
      <div v-for="item in items" :key="item.label" class="help-item">
        <span class="help-label mono">{{ item.label }}</span>
        <span class="help-body">{{ item.body }}</span>
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
.page-body {
  flex: 1;
  overflow: auto;
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.help-item {
  display: flex;
  gap: 10px;
  padding: 11px 13px;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 12.5px;
  line-height: 1.6;
}
.help-label {
  flex: 0 0 44px;
  color: var(--accent);
}
</style>
