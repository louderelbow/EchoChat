<template>
  <div style="height: 100%; width: 100%" v-if="isVisible">
    <el-select
      v-model="selectedGroupId"
      placeholder="选择要审核的群聊"
      style="width: 100%; margin-bottom: 10px"
      @change="loadMessages"
      filterable
    >
      <el-option
        v-for="g in groupList"
        :key="g.uuid"
        :label="g.name + '（' + g.uuid + '）'"
        :value="g.uuid"
      />
    </el-select>

    <el-scrollbar height="380px">
      <div v-if="!selectedGroupId" class="audit-empty">
        请选择群聊以查看聊天记录
      </div>
      <div v-else-if="messageList.length === 0" class="audit-empty">
        该群暂无聊天记录
      </div>
      <div v-else class="audit-list">
        <div v-for="(m, idx) in messageList" :key="idx" class="audit-item">
          <span class="audit-meta"
            >{{ m.created_at }} · {{ m.send_name }}</span
          >
          <div class="audit-content">
            <span v-if="m.type === undefined || m.type === 0">
              {{ m.content }}
            </span>
            <span v-else-if="m.type === 2">
              [文件] {{ m.file_name || m.url }}
            </span>
            <span v-else>[通话消息]</span>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script>
import { onMounted, reactive, toRefs } from "vue";
import { useStore } from "vuex";
import axios from "axios";
export default {
  name: "GroupAuditModal",
  props: {
    isVisible: false,
  },
  setup() {
    const store = useStore();
    const data = reactive({
      groupList: [],
      selectedGroupId: "",
      messageList: [],
    });

    onMounted(() => {
      getGroupInfoList();
    });

    const getGroupInfoList = async () => {
      try {
        const rsp = await axios.post(
          store.state.backendUrl + "/group/getGroupInfoList"
        );
        data.groupList = rsp.data.data || [];
      } catch (error) {
        console.log(error);
      }
    };

    const loadMessages = async (groupId) => {
      if (!groupId) {
        data.messageList = [];
        return;
      }
      try {
        const rsp = await axios.post(
          store.state.backendUrl + "/message/getGroupMessageList",
          { group_id: groupId }
        );
        data.messageList = rsp.data.data || [];
      } catch (error) {
        console.log(error);
        data.messageList = [];
      }
    };

    return {
      ...toRefs(data),
      getGroupInfoList,
      loadMessages,
    };
  },
};
</script>

<style scoped>
.audit-empty {
  color: #999;
  padding: 24px;
  text-align: center;
  font-size: 13px;
}
.audit-item {
  padding: 8px 10px;
  border-bottom: 1px solid #eee;
}
.audit-meta {
  color: #888;
  font-size: 12px;
}
.audit-content {
  margin-top: 4px;
  word-break: break-all;
  font-size: 14px;
}
</style>
