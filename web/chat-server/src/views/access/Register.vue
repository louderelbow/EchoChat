<template>
  <div class="register-wrap">
    <div
      class="register-window"
      :style="{
        boxShadow: `var(${'--el-box-shadow-dark'})`,
      }"
    >
      <div class="auth-head">
        <div class="auth-logo">E</div>
        <h2 class="register-item">注册</h2>
        <p class="auth-subtitle">创建账号，开启你的聊天之旅</p>
      </div>
      <el-form
        ref="formRef"
        :model="registerData"
        label-width="70px"
        class="demo-dynamic"
      >
        <el-form-item
          prop="nickname"
          label="昵称"
          :rules="[
            {
              required: true,
              message: '此项为必填项',
              trigger: 'blur',
            },
            {
              min: 3,
              max: 10,
              message: '昵称长度在 3 到 10 个字符',
              trigger: 'blur',
            },
          ]"
        >
          <el-input v-model="registerData.nickname" />
        </el-form-item>
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
          <el-input v-model="registerData.telephone" />
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
          <el-input type="password" v-model="registerData.password" />
        </el-form-item>
        <el-form-item
          prop="sms_code"
          label="验证码"
          :rules="[
            {
              required: true,
              message: '此项为必填项',
              trigger: 'blur',
            },
          ]"
        >
          <el-input v-model="registerData.sms_code" class="sms-code-input">
            <template #append>
              <el-button @click="sendSmsCode" class="sms-send-btn"
                >点击发送</el-button
              >
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      <div class="register-button-container">
        <el-button type="primary" class="register-btn" @click="handleRegister"
          >注册</el-button
        >
      </div>
      <div class="go-login-button-container">
        <button class="go-sms-login-btn" @click="handleSmsLogin">
          验证码登录
        </button>
        <button class="go-password-login-btn" @click="handleLogin">
          密码登录
        </button>
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
  name: "Register",
  setup() {
    const data = reactive({
      registerData: {
        telephone: "",
        password: "",
        nickname: "",
        sms_code: "",
      },
    });
    const router = useRouter();
    const store = useStore();
    const handleRegister = async () => {
      try {
        if (
          !data.registerData.nickname ||
          !data.registerData.telephone ||
          !data.registerData.password ||
          !data.registerData.sms_code
        ) {
          ElMessage.error("请填写完整注册信息。");
          return;
        }
        if (
          data.registerData.nickname.length < 3 ||
          data.registerData.nickname.length > 10
        ) {
          ElMessage.error("昵称长度在 3 到 10 个字符。");
          return;
        }
        if (!checkEmailValid()) {
          ElMessage.error("请输入有效的邮箱地址。");
          return;
        }
        const response = await axios.post(
          store.state.backendUrl + "/register",
          data.registerData
        ); // 发送POST请求
        if (response.data.code == 200) {
          ElMessage.success(response.data.message);
          console.log(response.data.message);
          // 查看avatar前缀有没有http
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
        } else {
          ElMessage.error(response.data.message);
          console.log(response.data.message);
        }
      } catch (error) {
        ElMessage.error(error);
        console.log(error);
      }
    };
    const checkEmailValid = () => {
      const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return regex.test(data.registerData.telephone);
    };

    const handleLogin = () => {
      router.push("/login");
    };

    const handleSmsLogin = () => {
      router.push("/smsLogin");
    };

    const sendSmsCode = async () => {
      if (
        !data.registerData.telephone ||
        !data.registerData.nickname ||
        !data.registerData.password
      ) {
        ElMessage.error("请填写完整注册信息。");
        return;
      }
      if (!checkEmailValid()) {
        ElMessage.error("请输入有效的邮箱地址。");
        return;
      }
      const req = {
        telephone: data.registerData.telephone,
      };
      const rsp = await axios.post(
        store.state.backendUrl + "/user/sendSmsCode",
        req
      );
      console.log(rsp);
      if (rsp.data.code == 200) {
        ElMessage.success(rsp.data.message);
      } else if (rsp.data.code == 400) {
        ElMessage.warning(rsp.data.message);
      } else {
        ElMessage.error(rsp.data.message);
      }
    };

    return {
      ...toRefs(data),
      router,
      handleRegister,
      handleLogin,
      handleSmsLogin,
      sendSmsCode,
    };
  },
};
</script>

<style>
.register-wrap {
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

.register-window {
  width: 420px;
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

.register-item {
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

.sms-code-input {
  width: 100%;
}

.sms-send-btn {
  background-color: var(--echo-primary-soft);
  border: none;
  color: var(--echo-primary-dark);
  font-weight: 600;
  border-radius: 0 var(--echo-radius-sm) var(--echo-radius-sm) 0;
}

.sms-send-btn:hover {
  background-color: var(--echo-primary-soft-2);
  color: var(--echo-primary-dark);
}

.register-button-container {
  display: flex;
  justify-content: center;
  margin-top: 22px;
  width: 100%;
}

.register-btn,
.register-btn:hover {
  width: 100%;
  height: 44px;
  background-color: var(--echo-primary);
  border: none;
  color: #ffffff;
  font-weight: 600;
  border-radius: var(--echo-radius-sm);
  font-size: 15px;
}

.register-btn:hover {
  background-color: var(--echo-primary-dark);
}

.el-alert {
  margin-top: 20px;
}

.go-login-button-container {
  display: flex;
  flex-direction: row-reverse;
  margin-top: 16px;
}

.go-sms-login-btn,
.go-password-login-btn {
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

.go-sms-login-btn:hover,
.go-password-login-btn:hover {
  background-color: var(--echo-primary-soft);
}
</style>