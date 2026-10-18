<template>
  <div class="chat-wrap">
    <div
      class="chat-window"
      :style="{
        boxShadow: `var(${'--el-box-shadow-dark'})`,
      }"
    >
      <el-container class="chat-window-container">
        <el-aside class="aside-container">
          <NavigationModal></NavigationModal>
          <ContactListModal></ContactListModal>
        </el-aside>
        <div class="owner-info-window">
          <div class="my-homepage-title"><h2>我的主页</h2></div>

          <p class="owner-prefix">用户id：{{ userInfo.uuid }}</p>
          <p class="owner-prefix">昵称：{{ userInfo.nickname }}</p>
          <p class="owner-prefix">邮箱：{{ userInfo.telephone }}</p>
          <p class="owner-prefix">手机号：{{ userInfo.email }}</p>
          <p class="owner-prefix">
            性别：{{ formatGender(userInfo.gender) }}
          </p>
          <p class="owner-prefix">生日：{{ userInfo.birthday }}</p>
          <p class="owner-prefix">个性签名：{{ userInfo.signature }}</p>
          <p class="owner-prefix">
            加入echo chat server的时间：{{ userInfo.created_at }}
          </p>
          <div class="owner-opt">
            <p class="owner-prefix">头像：</p>
            <img style="width: 40px; height: 40px" :src="userInfo.avatar" />
          </div>
        </div>
        <div class="edit-window">
          <el-button class="edit-btn" @click="showMyInfoModal">编辑</el-button>
        </div>
        <Modal :isVisible="isMyInfoModalVisible">
          <template v-slot:header>
            <div class="modal-header">
              <div class="modal-quit-btn-container">
                <button class="modal-quit-btn" @click="quitMyInfoModal">
                  <el-icon><Close /></el-icon>
                </button>
              </div>
              <div class="modal-header-title">
                <h3>修改主页</h3>
              </div>
            </div>
          </template>
          <template v-slot:body>
            <el-scrollbar
              max-height="300px"
              style="
                width: 400px;
                display: flex;
                align-items: center;
                justify-content: center;
                margin-top: 20px;
              "
            >
              <div class="modal-body">
                <el-form ref="formRef" :model="updateInfo" label-width="70px">
                  <el-form-item
                    prop="nickname"
                    label="昵称"
                    :rules="[
                      {
                        min: 3,
                        max: 10,
                        message: '昵称长度在 3 到 10 个字符',
                        trigger: 'blur',
                      },
                    ]"
                  >
                    <el-input
                      v-model="updateInfo.nickname"
                      placeholder="选填"
                    />
                  </el-form-item>
                  <el-form-item prop="email" label="手机号">
                    <el-input
                      v-model="updateInfo.email"
                      placeholder="选填，请输入手机号"
                    />
                  </el-form-item>
                  <el-form-item prop="gender" label="性别">
                    <el-radio-group v-model="updateInfo.gender">
                      <el-radio :value="0">未知</el-radio>
                      <el-radio :value="1">男</el-radio>
                      <el-radio :value="2">女</el-radio>
                    </el-radio-group>
                  </el-form-item>
                  <el-form-item prop="birthday" label="生日">
                    <el-input
                      v-model="updateInfo.birthday"
                      placeholder="选填，格式为2024.1.1"
                    />
                  </el-form-item>
                  <el-form-item prop="signature" label="个性签名">
                    <el-input
                      v-model="updateInfo.signature"
                      placeholder="选填"
                    />
                  </el-form-item>
                  <el-form-item prop="avatar" label="头像">
                    <el-upload
                      v-model:file-list="fileList"
                      ref="uploadRef"
                      :auto-upload="false"
                      :action="uploadPath"
                      :on-success="handleUploadSuccess"
                      :before-upload="beforeFileUpload"
                    >
                      <template #trigger>
                        <el-button
                          class="echo-soft-btn"
                          >上传图片</el-button
                        >
                      </template>
                    </el-upload>
                  </el-form-item>
                </el-form>
              </div>
            </el-scrollbar>
          </template>
          <template v-slot:footer>
            <div class="modal-footer">
              <el-button class="modal-close-btn" @click="closeMyInfoModal">
                完成
              </el-button>
            </div>
          </template>
        </Modal>
      </el-container>
    </div>
  </div>
</template>

<script>
import { reactive, toRefs, onMounted, ref } from "vue";
import { useStore } from "vuex";
import axios from "axios";
import { useRouter } from "vue-router";
import Modal from "@/components/Modal.vue";
import { checkTelephoneValid } from "@/assets/js/valid.js";
import { generateString } from "@/assets/js/random.js";
import SmallModal from "@/components/SmallModal.vue";
import NavigationModal from "@/components/NavigationModal.vue";
import ContactListModal from "@/components/ContactListModal.vue";
import { ElMessage } from "element-plus";
export default {
  name: "OwnInfo",
  components: {
    Modal,
    SmallModal,
    ContactListModal,
    NavigationModal,
  },
  setup() {
    const router = useRouter();
    const store = useStore();
    const data = reactive({
      userInfo: store.state.userInfo,
      updateInfo: {
        uuid: "",
        nickname: "",
        email: "",
        gender: 0,
        birthday: "",
        signature: "",
        avatar: "",
      },
      isMyInfoModalVisible: false,
      ownListReq: {
        owner_id: "",
      },
      uploadRef: null,
      uploadPath: store.state.backendUrl + "/message/uploadAvatar",
      fileList: [],
      cnt: 0,
    });
    const showMyInfoModal = () => {
      data.updateInfo.gender = data.userInfo.gender || 0;
      data.isMyInfoModalVisible = true;
    };
    // 提交个人资料更新：关弹窗 → 调接口 → 刷新
    const submitMyInfo = async () => {
      data.isMyInfoModalVisible = false;
      data.cnt = 0;
      data.updateInfo.uuid = data.userInfo.uuid;
      store.commit("setUserInfo", data.userInfo);
      try {
        const rsp = await axios.post(
          store.state.backendUrl + "/user/updateUserInfo",
          data.updateInfo
        );
        console.log(rsp);
        if (rsp.data.code == 200) {
          ElMessage.success(rsp.data.message);
        } else {
          ElMessage.error(rsp.data.message);
        }
      } catch (error) {
        console.log(error);
      }
      router.go(0);
    };

    const closeMyInfoModal = async () => {
      console.log(data.fileList);
      if (
        data.updateInfo.nickname == "" &&
        data.fileList.length == 0 &&
        data.updateInfo.email == "" &&
        data.updateInfo.birthday == "" &&
        data.updateInfo.signature == "" &&
        data.updateInfo.gender == data.userInfo.gender
      ) {
        ElMessage("请至少修改一项");
        return;
      }
      if (data.updateInfo.nickname != "") {
        if (
          data.updateInfo.nickname.length < 3 ||
          data.updateInfo.nickname.length > 10
        ) {
          return;
        }
        data.userInfo.nickname = data.updateInfo.nickname;
      }
      if (data.updateInfo.email != "") {
        if (!checkTelephoneValid(data.updateInfo.email)) {
          ElMessage("请输入有效的手机号。");
          return;
        }
        data.userInfo.email = data.updateInfo.email;
      }
      if (data.updateInfo.birthday != "") {
        data.userInfo.birthday = data.updateInfo.birthday;
      }
      if (data.updateInfo.signature != "") {
        data.userInfo.signature = data.updateInfo.signature;
      }
      if (data.updateInfo.gender != data.userInfo.gender) {
        data.userInfo.gender = data.updateInfo.gender;
      }

      // 有选头像：先发起上传，等上传成功后在回调里再提交资料并刷新
      // （不能在这里就关弹窗 / 刷新，否则会中断正在上传的请求）
      if (data.fileList.length != 0) {
        data.updateInfo.avatar = "/static/avatars/" + data.fileList[0].name;
        console.log(data.updateInfo.avatar);
        data.userInfo.avatar = store.state.backendUrl + data.updateInfo.avatar;
        store.commit("setUserInfo", data.userInfo);
        data.uploadRef.submit();
        return;
      }

      await submitMyInfo();
    };
    const quitMyInfoModal = () => {
      data.isMyInfoModalVisible = false;
      data.fileList = [];
      data.cnt = 0;
    };
    const handleUploadSuccess = async () => {
      ElMessage.success("头像上传成功");
      data.fileList = [];
      await submitMyInfo();
    };
    const beforeFileUpload = (file) => {
      console.log("上传前file====>", file);
      console.log(data.fileList);
      console.log(file);
      if (data.fileList.length > 1) {
        ElMessage.error("只能上传一张头像");
        return false;
      }
      const isLt50M = file.size / 1024 / 1024 < 50;
      if (!isLt50M) {
        ElMessage.error("上传头像图片大小不能超过 50MB!");
        return false;
      }
    };
    const getFileExtension = (filename) => {
      const parts = filename.split(".");
      return parts.length > 1 ? parts.pop() : "";
    };
    // 性别：0.未知，1.男，2.女
    const formatGender = (g) => {
      if (g === 1) return "男";
      if (g === 2) return "女";
      return "未知";
    };

    
    return {
      ...toRefs(data),
      router,
      showMyInfoModal,
      closeMyInfoModal,
      quitMyInfoModal,
      handleUploadSuccess,
      beforeFileUpload,
      formatGender,
    };
  },
};
</script>

<style scoped>
.owner-info-window {
  width: 84%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.owner-prefix {
  font-family: Arial, Helvetica, sans-serif;
  margin: 6px;
}

.owner-opt {
  margin: 6px;
  display: flex;
  flex-direction: row;
}

.edit-window {
  width: 10%;
  display: flex;
  flex-direction: column-reverse;
}

h3 {
  font-family: Arial, Helvetica, sans-serif;
  color: rgb(69, 69, 68);
}

.modal-quit-btn-container {
  height: 30%;
  width: 100%;
  display: flex;
  flex-direction: row-reverse;
}

.modal-quit-btn {
  background-color: rgba(255, 255, 255, 0);
  color: rgb(229, 25, 25);
  padding: 15px;
  border: none;
  cursor: pointer;
  position: fixed;
  justify-content: center;
  align-items: center;
}

.modal-header {
  height: 20%;
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.modal-body {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.modal-footer {
  height: 20%;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal-header-title {
  height: 70%;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

h2 {
  margin-bottom: 20px;
  font-family: Arial, Helvetica, sans-serif;
}

.el-menu {
  background-color: rgb(252, 210.9, 210.9);
  width: 101%;
}

.el-menu-item {
  background-color: transparent;
  height: 46px;
}

/* ===== 现代清新主题覆盖（仅样式，逻辑不变） ===== */
.owner-info-window {
  width: 100%;
  flex: 1;
  min-width: 0;
  padding: 32px 40px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: flex-start;
  overflow-y: auto;
  box-sizing: border-box;
}

.my-homepage-title h2 {
  margin: 0 0 22px;
  font-size: 22px;
  color: var(--echo-text);
  font-family: var(--echo-font);
}

.owner-prefix {
  font-family: var(--echo-font);
  font-size: 14px;
  color: var(--echo-text-2);
  margin: 5px 0;
}

.owner-opt {
  margin: 14px 0 0;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.owner-opt img {
  border-radius: 50%;
  box-shadow: var(--echo-shadow-sm);
}

.edit-window {
  width: 96px;
  flex: 0 0 auto;
  padding: 32px 20px 32px 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  box-sizing: border-box;
}

.el-menu {
  background-color: transparent;
  width: 100%;
}

.modal-header,
.modal-body,
.modal-footer,
.modal-quit-btn-container {
  height: auto;
}

.modal-quit-btn {
  position: static;
}
</style>