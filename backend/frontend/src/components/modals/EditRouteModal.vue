<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <div class="modal-header">
        <h2>Edit Route</h2>
        <button class="btn btn-icon" @click="$emit('close')">×</button>
      </div>
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label>Name</label>
          <input type="text" v-model="form.name" required>
        </div>
        <div class="form-group">
          <label>Domain</label>
          <input type="text" v-model="form.domain" required>
        </div>
        <div class="form-group">
          <label>Target</label>
          <input type="text" v-model="form.target" required>
        </div>
        <div class="checkbox-group">
          <input type="checkbox" id="edit-enabled" v-model="form.enabled">
          <label for="edit-enabled" style="margin: 0;">Enabled</label>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-icon" @click="$emit('close')">Cancel</button>
          <button type="submit" class="btn btn-primary">Save</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
export default {
  name: 'EditRouteModal',
  props: {
    route: {
      type: Object,
      required: true,
    },
  },
  emits: ['close', 'save'],
  data() {
    return {
      form: { ...this.route },
    }
  },
  watch: {
    route: {
      handler(newRoute) {
        this.form = { ...newRoute }
      },
      immediate: true,
    },
  },
  methods: {
    handleSubmit() {
      this.$emit('save', this.form)
    },
  },
}
</script>
