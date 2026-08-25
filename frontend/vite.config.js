import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const rootDir = dirname(fileURLToPath(import.meta.url)) // 本项目 frontend/ 目录
const dfclientkitDir = resolve(rootDir, '../../../../dfclientkit')

export default defineConfig({
  plugins: [vue()],
  optimizeDeps: {
    // df-ui-shell 是纯源码分发的包（.vue 文件不预编译），
    // esbuild 的依赖预打包不认识 .vue 语法，必须排除掉。
    exclude: ['@dongfang/df-ui-shell'],
  },
  server: {
    fs: {
      // dfclientkit 在项目目录树之外（本地 file: 依赖），Vite 默认的
      // 文件系统访问白名单只到项目根目录，这里需要显式放行。
      allow: [rootDir, dfclientkitDir],
    },
  },
})
