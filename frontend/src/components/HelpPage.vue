<script setup>
const items = [
  { label: '权限', body: '在 GitHub 生成 Personal Access Token，需勾选 repo 权限；若要自动建仓库还需 public_repo。' },
  { label: '格式', body: '支持导入 .md / .txt 文件，.txt 会以 .md 提交，文件名作为标题。' },
  { label: '发布', body: '单账号、单仓库；文件已存在则更新覆盖，每篇一次提交。' },
  { label: '重试', body: '可设置发布间隔与失败重试次数；遇 API 限流会按 Retry-After 等待重试。' },
  { label: '配置', body: '设置（含 Token）保存在本机的配置文件中，下次打开自动载入。' },
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
