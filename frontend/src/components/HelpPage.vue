<script setup>
const items = [
  { label: '模板', body: '「内容设置 → 模板」里写标题和正文模板，用 {关键词}、{变量1} 这样的占位符标出每篇要变的地方，其余文字原样保留。' },
  { label: '占位符', body: '{关键词} 取关键词库里的当前一条；{变量1}~{变量5} 各自从对应词库随机抽一行；{英文=6}/{小写=4}/{数字=3} 生成指定位数的随机串；{日期}、{时间} 用生成时刻。同一个占位符出现多次会各自重新抽取，写错的占位符原样留在结果里，方便发现笔误。' },
  { label: '词库', body: '「词库」页里一行一条。关键词可以选顺序轮转还是随机抽取，也可以选生成时是否在字符之间加空格。' },
  { label: 'txt 配置', body: '模板和词库都以 txt 文件存在本机，点右侧「打开素材目录」可以直接用文本编辑器批量编辑；每个输入框旁的「从文件导入」也能一次灌进一整个词库。' },
  { label: '生成', body: '设好篇数点「生成」，结果在「预览」页逐条展开查看。可选去除正文里的重复行、或只保留含中文的行。' },
  { label: '导出', body: '点「导出 .md」选一个文件夹，每篇存成一个 Markdown 文件，标题作为文件名（非法字符会被替换，重名自动加序号）。' },
]
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">
        <span class="page-title-main">使用说明</span>
        <span class="page-title-sub">占位符语法、词库格式与生成导出流程</span>
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
  flex: 0 0 62px;
  color: var(--accent);
}
</style>
