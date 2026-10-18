<template>
  <div class="login-wrap">
    <div
      class="login-window"
      :style="{
        boxShadow: `var(${'--el-box-shadow-dark'})`,
      }"
    >
      <div class="auth-head">
        <div class="auth-logo">E</div>
        <h2 class="login-item">登录</h2>
        <p class="auth-subtitle">欢迎回来，请登录你的账号</p>
      </div>
      <el-form
        :model="loginData"
        label-width="70px"
        class="demo-dynamic"
      >
        <el-form-item
          prop="telephone"
          label="邮箱"
          :rules="[
            {
              required: true,
              message: '此项为必填项',
              trigger: 'blur',
            },
          ]"
        >
          <el-input v-model="loginData.telephone" />
        </el-form-item>
        <el-form-item
          prop="password"
          label="密码"
          :rules="[
            {
              required: true,
              message: '此项为必填项',
              trigger: 'blur',
            },
          ]"
        >
          <el-input type="password" v-model="loginData.password" />
        </el-form-item>
      </el-form>
      <div class="login-button-container">
        <el-button type="primary" class="login-btn" @click="handleLogin"
          >登录</el-button
        >
      </div>

      <div class="go-register-button-container">
        <button class="go-register-btn" @click="handleRegister">注册</button>
        <button class="go-sms-btn" @click="handleSmsLogin">验证码登录</button>
      </div>
    </div>
  </div>
</template>

<script>
import { reactive, toRefs } from "vue";
import axios from "axios";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { useStore } from "vuex";
export default {
  name: "Login",
  setup() {
    const data = reactive({
      loginData: {
        telephone: "",
        password: "",
      },
    });
    const router = useRouter();
    const store = useStore();
    const handleLogin = async () => {
      try {
        if (!data.loginData.telephone || !data.loginData.password) {
          ElMessage.error("请填写完整登录信息。");
          return;
        }
        if (!checkEmailValid()) {
          ElMessage.error("请输入有效的邮箱地址。");
          return;
        }
	console.log(store.state.backendUrl, store.state.wsUrl);
        const response = await axios.post(
          store.state.backendUrl + "/login",
          data.loginData
        );
        console.log(response);
        if (response.data.code == 200) {
          if (response.data.data.status == 1) {
            ElMessage.error("该账号已被封禁，请联系管理员。");
            return;
          }
          try {
            ElMessage.success(response.data.message);
            if (!response.data.data.avatar.startsWith("http")) {
              response.data.data.avatar =
                store.state.backendUrl + response.data.data.avatar;
            }
            store.commit("setUserInfo", response.data.data);
            store.commit("setToken", response.data.token);
            // 准备创建websocket连接
            const wsUrl =
              store.state.wsUrl +
              "/wss?client_id=" +
              response.data.data.uuid +
              "&token=" +
              (response.data.token || "");
            console.log(wsUrl);
            store.state.socket = new WebSocket(wsUrl);
            store.state.socket.onopen = () => {
              console.log("WebSocket连接已打开");
            };
            store.state.socket.onmessage = (message) => {
              console.log("收到消息：", message.data);
            };
            store.state.socket.onclose = () => {
              console.log("WebSocket连接已关闭");
            };
            store.state.socket.onerror = () => {
              console.log("WebSocket连接发生错误");
            };
            router.push("/chat/sessionlist");
          } catch (error) {
            console.log(error);
          }
        } else {
          ElMessage.error(response.data.message);
        }
      } catch (error) {
        ElMessage.error(error);
      }
    };
    const checkEmailValid = () => {
      const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return regex.test(data.loginData.telephone);
    };
    const handleRegister = () => {
      router.push("/register");
    };
    const handleSmsLogin = () => {
      router.push("/smsLogin");
    };

    return {
      ...toRefs(data),
      router,
      handleLogin,
      handleRegister,
      handleSmsLogin,
    };
  },
};
</script>

<style>
.login-wrap {
  height: 100vh;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--echo-bg-gradient);
  background-attachment: fixed;
  padding: 24px;
  box-sizing: border-box;
}

.login-window {
  width: 400px;
  max-width: 100%;
  background-color: rgba(255, 255, 255, 0.94);
  padding: 38px 44px 32px;
  border-radius: var(--echo-radius-lg);
  border: 1px solid rgba(255, 255, 255, 0.85);
  box-shadow: var(--echo-shadow-lg) !important;
  box-sizing: border-box;
}

.auth-head {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 24px;
}

.auth-logo {
  width: 52px;
  height: 52px;
  border-radius: 16px;
  background: linear-gradient(135deg, var(--echo-primary), var(--echo-accent));
  color: #fff;
  font-size: 26px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--echo-shadow);
  margin-bottom: 14px;
}

.login-item {
  text-align: center;
  margin: 0 0 6px;
  color: var(--echo-text);
  font-size: 22px;
  font-family: var(--echo-font);
}

.auth-subtitle {
  margin: 0;
  color: var(--echo-text-2);
  font-size: 13px;
}

.login-button-container {
  display: flex;
  justify-content: center;
  margin-top: 22px;
  width: 100%;
}

.login-btn,
.login-btn:hover {
  width: 100%;
  height: 44px;
  background-color: var(--echo-primary);
  border: none;
  color: #ffffff;
  font-weight: 600;
  border-radius: var(--echo-radius-sm);
  font-size: 15px;
}

.login-btn:hover {
  background-color: var(--echo-primary-dark);
}

.go-register-button-container {
  display: flex;
  flex-direction: row-reverse;
  margin-top: 16px;
}

.go-register-btn,
.go-sms-btn {
  background-color: transparent;
  border: none;
  cursor: pointer;
  color: var(--echo-primary-dark);
  font-weight: 600;
  font-size: 13px;
  padding: 4px 6px;
  margin-left: 6px;
  border-radius: 6px;
}

.go-register-btn:hover,
.go-sms-btn:hover {
  background-color: var(--echo-primary-soft);
}

.el-alert {
  margin-top: 20px;
}
</style>
