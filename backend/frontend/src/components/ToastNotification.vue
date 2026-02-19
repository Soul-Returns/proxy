<template>
  <Transition name="toast">
    <div v-if="show" :class="['toast', 'toast-' + type]">
      {{ message }}
    </div>
  </Transition>
</template>

<script>
export default {
  name: 'ToastNotification',
  props: {
    show: Boolean,
    message: String,
    type: {
      type: String,
      default: 'success',
    },
  },
}
</script>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(1rem);
}
</style>
