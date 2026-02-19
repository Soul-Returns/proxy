<template>
  <div class="modal-overlay">
    <div class="modal loading-modal">
      <div class="loading-spinner"></div>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'LoadingModal',
  props: {
    message: {
      type: String,
      default: 'Loading...',
    },
  },
}
</script>
