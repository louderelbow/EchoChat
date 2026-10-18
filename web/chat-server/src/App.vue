<template>
  <router-view />
</template>

<script>
import { onMounted, watch } from "vue";
import { useStore } from "vuex";
import { useRouter } from "vue-router";
import { ElNotification } from "element-plus";
import axios from "axios";
export default {
  name: "App",
  setup() {
    const store = useStore();
    const router = useRouter();

    // 全局通话信令处理：无论当前在哪个页面都能收到来电提示
    // 说明：用 addEventListener 挂载，避免与 ContactChat 里的 socket.onmessage 互相覆盖
    const handleCallSignal = (jsonMessage) => {
      try {
        const message = JSON.parse(jsonMessage.data);
        if (Number(message.type) !== 3 || !message.av_data) {
          return;
        }
        const avData = JSON.parse(message.av_data);
        const myId = store.state.userInfo.uuid;
        if (avData.messageId === "PROXY" && avData.type === "start_call") {
          // 群聊来电：receive_id 是群 id（G 开头），所有群成员都应收到提示；
          // 单聊来电：只处理发给自己的
          const isGroup =
            message.receive_id && message.receive_id[0] === "G";
          if (!isGroup && myId && message.receive_id && message.receive_id !== myId) {
            return;
          }
          store.commit("setPendingCall", {
            fromId: message.send_id,
            fromName: message.send_name,
            receiveId: message.receive_id,
            isGroup,
            // 点击提示跳转的目标：群聊进群、单聊进对方的会话
            targetId: isGroup ? message.receive_id : message.send_id,
            ts: Date.now(),
          });
        } else if (
          (avData.messageId === "PROXY" && avData.type === "reject_call") ||
          avData.messageId === "PEER_LEAVE"
        ) {
          store.commit("clearPendingCall");
        }
      } catch (error) {
        // 非信令消息，忽略
      }
    };

    // 建立/重建 WebSocket 连接；断开后 3 秒自动重连，避免网络抖动导致静默失联
    let reconnectTimer = null;
    const scheduleReconnect = () => {
      if (reconnectTimer) return;
      if (!store.state.userInfo.uuid) return;
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null;
        connectSocket();
      }, 3000);
    };
    const connectSocket = () => {
      const uuid = store.state.userInfo.uuid;
      if (!uuid) return;
      const wsUrl =
        store.state.wsUrl +
        "/wss?client_id=" +
        uuid +
        "&token=" +
        (store.state.token || "");
      console.log(wsUrl);
      const socket = new WebSocket(wsUrl);
      store.state.socket = socket;
      socket.onopen = () => {
        console.log("WebSocket连接已打开");
        if (reconnectTimer) {
          clearTimeout(reconnectTimer);
          reconnectTimer = null;
        }
      };
      socket.onmessage = (message) => {
        console.log("收到消息：", message.data);
      };
      socket.onclose = () => {
        console.log("WebSocket连接已关闭，将在 3 秒后重连");
        scheduleReconnect();
      };
      socket.onerror = () => {
        console.log("WebSocket连接发生错误");
      };
    };

    // socket 可能在 App / 登录页 / 聊天页被创建或替换，监听其变化以挂载全局处理
    watch(
      () => store.state.socket,
      (socket) => {
        if (socket && socket.addEventListener) {
          socket.addEventListener("message", handleCallSignal);
          // 保证任何位置创建的 socket 断开后都能自动重连
          socket.onclose = () => scheduleReconnect();
        }
      },
      { immediate: true }
    );

    // 来电提示：由全局 pendingCall 驱动，常驻可关闭、点击可进入会话接听
    let callNotifyInstance = null;
    watch(
      () => store.state.pendingCall,
      (call) => {
        if (callNotifyInstance) {
          callNotifyInstance.close();
          callNotifyInstance = null;
        }
        if (call) {
          const target = call.isGroup ? call.receiveId : call.fromId;
          callNotifyInstance = ElNotification({
            title: "消息提示",
            message: call.isGroup
              ? `群聊收到来自${call.fromName || "对方"}的通话请求，点击前往接听`
              : `收到一条来自${call.fromName || "对方"}的通话请求，点击前往接听`,
            type: "warning",
            duration: 0,
            showClose: true,
            onClick: () => {
              router.push("/chat/" + target);
            },
          });
        }
      }
    );

    const getUserInfo = async () => {
      try {
        const req = {
          uuid: store.state.userInfo.uuid,
        };
        const rsp = await axios.post(
          store.state.backendUrl + "/user/getUserInfo",
          req
        );
        if (rsp.data.code == 200) {
          if (rsp.data.data.avatar && !rsp.data.data.avatar.startsWith("http")) {
            rsp.data.data.avatar = store.state.backendUrl + rsp.data.data.avatar;
          }
          store.commit("setUserInfo", rsp.data.data);
        } else {
          console.error(rsp.data.message);
        }
        console.log(rsp);
      } catch (error) {
        console.log(error);
      }
    };
    const logout = async () => {
      store.commit("cleanUserInfo");
      const req = {
        owner_id: data.userInfo.uuid,
      };
      const rsp = await axios.post(
        store.state.backendUrl + "/user/wsLogout",
        req
      );
      if (rsp.data.code == 200) {
        router.push("/login");
        ElMessage.success("账号被封禁，退出登录");
      } else {
        ElMessage.error(rsp.data.message);
      }
    };
    onMounted(() => {
      if (store.state.userInfo.uuid) {
        getUserInfo();
        if (store.state.userInfo.status == 1) {
          logout();
        }
        connectSocket();
      }
    });
  },
};
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box; /* 推荐使用，以确保布局计算的一致性 */
}
</style>