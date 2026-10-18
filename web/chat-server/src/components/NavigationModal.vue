<template>
  <div class="navigation-bar">
    <div class="up-bar">
      <button class="avatar-btn">
        <el-image
          class="nav-avatar"
          :src="userInfo.avatar"
          :preview-src-list="[userInfo.avatar]"
          :preview-teleported="true"
          fit="cover"
        />
      </button>
    </div>
    <div class="middle-bar">
      <el-tooltip
        effect="customized"
        content="会话聊天"
        placement="left"
        hide-after="0"
        enterable="false"
      >
        <button
          class="icon-btn"
          :class="{ 'is-active': activeNav === 'session' }"
          @click="handleToSessionList"
        >
          <el-icon>
            <ChatRound />
          </el-icon>
        </button>
      </el-tooltip>
      <el-tooltip
        effect="customized"
        content="通讯录管理"
        placement="left"
        hide-after="0"
        enterable="false"
      >
        <el-badge
          :value="newContactCount"
          :hidden="!newContactCount"
          :max="99"
          class="nav-badge"
        >
          <button
            class="icon-btn"
            :class="{ 'is-active': activeNav === 'contact' }"
            @click="handleToContactList"
          >
            <el-icon>
              <User />
            </el-icon>
          </button>
        </el-badge>
      </el-tooltip>
      <el-tooltip
        effect="customized"
        content="朋友圈"
        placement="left"
        hide-after="0"
        enterable="false"
      >
        <button class="icon-btn" @click="notImplemented">
          <el-icon>
            <Share />
          </el-icon>
        </button>
      </el-tooltip>
      <el-tooltip
        effect="customized"
        content="我的收藏"
        placement="left"
        hide-after="0"
        enterable="false"
      >
        <button class="icon-btn" @click="notImplemented">
          <el-icon>
            <Star />
          </el-icon>
        </button>
      </el-tooltip>
      <el-tooltip
        effect="customized"
        content="搜索"
        placement="left"
        hide-after="0"
        enterable="false"
      >
        <button class="icon-btn" @click="notImplemented">
          <el-icon>
            <Search />
          </el-icon>
        </button>
      </el-tooltip>
    </div>
    <div class="down-bar">
      <el-tooltip
        effect="customized"
        content="设置"
        placement="left"
        hide-after="0"
        enterable="false"
      >
        <el-dropdown trigger="click" placement="right">
          <button class="icon-btn">
            <el-icon>
              <Setting />
            </el-icon>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-if="userInfo.is_admin == 1"
                @click="handleToManager"
              >
                管理员模式
              </el-dropdown-item>
              <el-dropdown-item @click="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-tooltip>
      <el-tooltip
        effect="customized"
        content="我的主页"
        placement="left"
        hide-after="0"
        enterable="false"
      >
        <button
          class="icon-btn"
          :class="{ 'is-active': activeNav === 'own' }"
          @click="handleToOwnInfo"
        >
          <el-icon><HomeFilled /></el-icon>
        </button>
      </el-tooltip>
    </div>
  </div>
</template>

<script>
import { useRoute, useRouter } from "vue-router";
import { useStore } from "vuex";
import { ElMessage } from "element-plus";
import { computed, reactive, toRefs, onMounted, onUnmounted } from "vue";
import axios from "axios";
export default {
  name: "NavigationModal",
  setup() {
    const router = useRouter();
    const route = useRoute();
    const store = useStore();
    const data = reactive({
      userInfo: store.state.userInfo,
      // 待处理申请数量（好友申请 + 我作为群主的加群申请），用于通讯录入口的红点
      newContactCount: 0,
    });

    // 当前所在导航项：session 会话 / contact 通讯录 / own 我的主页
    const activeNav = computed(() => {
      const p = route.path;
      if (p === "/chat/contactlist") return "contact";
      if (p === "/chat/owninfo") return "own";
      if (p === "/chat/sessionlist") return "session";
      if (/^\/chat\/[^/]+$/.test(p)) return "session"; // 具体的聊天页也归到会话
      return "";
    });

    // 尚未实现的功能统一提示
    const notImplemented = () => {
      ElMessage("功能开发中，敬请期待");
    };

    const handleToContactList = () => {
      router.push("/chat/contactlist");
    };

    const handleToSessionList = () => {
      router.push("/chat/sessionlist");
    };

    const handleToManager = () => {
      console.log(data.userInfo);
      router.push("/manager");
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
        ElMessage.success(rsp.data.message);
      } else {
        ElMessage.error(rsp.data.message);
      }
    };
    const handleToOwnInfo = () => {
      router.push("/chat/owninfo");
    };

    // 拉取待处理申请数量用于红点，定时轮询保证新申请能及时提示
    let countTimer = null;
    const loadNewContactCount = async () => {
      try {
        const uuid = store.state.userInfo.uuid;
        if (!uuid) return;
        const rsp = await axios.post(
          store.state.backendUrl + "/contact/getNewContactCount",
          { owner_id: uuid }
        );
        if (rsp.data.code == 200) {
          data.newContactCount = Number(rsp.data.data || 0);
        }
      } catch (error) {
        console.error(error);
      }
    };
    onMounted(() => {
      loadNewContactCount();
      countTimer = setInterval(loadNewContactCount, 20000);
    });
    onUnmounted(() => {
      if (countTimer) {
        clearInterval(countTimer);
        countTimer = null;
      }
    });
    return {
      ...toRefs(data),
      router,
      activeNav,
      notImplemented,
      handleToContactList,
      handleToSessionList,
      handleToOwnInfo,
      logout,
      handleToManager,
    };
  },
};
</script>

<style scoped>
/* 红点：让徽标包裹层不影响图标按钮的原有布局 */
.nav-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
</style>