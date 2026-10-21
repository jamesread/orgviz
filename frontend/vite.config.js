import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const backendPort = process.env.PORT || '8080'
const backendTarget = 'http://127.0.0.1:' + backendPort

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'node'
  },
  server: {
    proxy: {
      '/api': {
        target: backendTarget,
        changeOrigin: true,
        secure: false
      },
      '/avatars': {
        target: backendTarget,
        changeOrigin: true,
        secure: false
      }
    }
  }
})
