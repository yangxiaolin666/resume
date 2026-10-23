import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  base: './',
  plugins: [
    vue(),
    vueDevTools(),
    AutoImport({
      resolvers: [ElementPlusResolver()]
    }),
    Components({
      resolvers: [ElementPlusResolver()]
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    // 开发环境把 /api-backend 转发到 Go 后端。
    // 这样浏览器只访问 5173（同源），不会再出现跨域预检失败、
    // 8081 端口被浏览器/插件拦截、或 HTTPS 预览下的混合内容问题。
    proxy: {
      '/api-backend': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
        ws: true, // 支持 WebSocket 聊天
        rewrite: (path) => path.replace(/^\/api-backend/, '')
      }
    }
  }
})
