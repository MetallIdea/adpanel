import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import zipPack from "vite-plugin-zip-pack";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), zipPack({
    outDir:'../dist/',
    outFileName:'frontend.zip'
  })],
 server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    }
  },
})
