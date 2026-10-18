<template>

  <div class="modal-overlay" v-if="isVisible">
    <div class="modal-content">
      <slot name="header">
      </slot>
      <slot name="body">
      </slot>
      <slot name="footer">
      </slot>
    </div>
  </div>
</template>

<script>
export default {
  name: "Modal",
  props: {
    isVisible: false,
  },
};
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(28, 44, 66, 0.42);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
  padding: 24px;
  box-sizing: border-box;
}

.modal-content {
  background: #fff;
  border-radius: var(--echo-radius-lg);
  width: 400px;
  max-width: 100%;
  max-height: 86vh;
  padding: 16px 22px 20px;
  box-shadow: var(--echo-shadow-lg);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
}
</style>
