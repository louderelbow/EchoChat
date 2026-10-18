const { defineConfig } = require('@vue/cli-service')
const fs = require('fs')
const path = require('path')

module.exports = defineConfig({
  transpileDependencies: true,
  lintOnSave: false,
  devServer: {
    host: '0.0.0.0',

    /* ---------- 当前启用：本地 HTTPS ----------
     * 证书需先用 mkcert 生成到 src/assets/cert/ 下（见答复步骤）。
     * HTTPS 约定端口为 443，浏览器访问 https://127.0.0.1（443 可省略端口号）。
     */
    port: 443,
    https: {
      cert: fs.readFileSync(path.join(__dirname, 'src/assets/cert/127.0.0.1.pem')),
      key: fs.readFileSync(path.join(__dirname, 'src/assets/cert/127.0.0.1-key.pem')),
    },

    /* ---------- 如需切回本地 HTTP，注释掉上面两段、启用下面这段 ----------
     * 单机调试且不需要安全上下文时用这版，端口 8080。
     */
    // port: 8080,
  }
})
