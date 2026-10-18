import { createStore } from 'vuex'

// 后端地址：优先取环境变量（Vue CLI 的 VUE_APP_* 会在构建时注入），
// 未配置时回落到本地开发默认值（HTTPS/WSS，与后端 TLS 对齐）。
// 部署时在 web/chat-server/.env.production 里设置：
//   VUE_APP_BACKEND_URL=https://your-domain
//   VUE_APP_WS_URL=wss://your-domain
const BACKEND_URL = process.env.VUE_APP_BACKEND_URL || 'https://127.0.0.1:8888'
const WS_URL = process.env.VUE_APP_WS_URL || 'wss://127.0.0.1:8888'

export default createStore({
  state: {
    // web服务器地址（已切换为 HTTPS / WSS，与后端 TLS 保持一致）
    backendUrl: BACKEND_URL,
    wsUrl: WS_URL,
    // 如需切回明文 HTTP，改回：
    // backendUrl: 'http://127.0.0.1:8888',
    // wsUrl: 'ws://127.0.0.1:8888',
    // 信令服务器地址
    // signalUrl: 'wss://127.0.0.1:8001',
    userInfo: (sessionStorage.getItem('userInfo') && JSON.parse(sessionStorage.getItem('userInfo'))) || {},
    // 登录令牌（鉴权开启时随请求头与 WS 携带）
    token: sessionStorage.getItem('token') || '',
    socket: null,
    // 收到但尚未接听的通话请求（全局，供任意页面提示后进入会话接听）
    pendingCall: null,
  },
  getters: {
  },
  mutations: {
    setUserInfo(state, userInfo) {
      state.userInfo = userInfo;
      sessionStorage.setItem('userInfo', JSON.stringify(userInfo));
    },
    setToken(state, token) {
      state.token = token || '';
      sessionStorage.setItem('token', state.token);
    },
    cleanUserInfo(state) {
      state.userInfo = {};
      sessionStorage.removeItem('userInfo');
      sessionStorage.removeItem('token');
      state.token = '';
      state.pendingCall = null;
    },
    setPendingCall(state, call) {
      state.pendingCall = call;
    },
    clearPendingCall(state) {
      state.pendingCall = null;
    }
  },
  actions: {
  },
  modules: {
  }
})
