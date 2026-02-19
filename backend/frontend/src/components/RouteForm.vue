<template>
  <div class="card">
    <h2>Add New Route</h2>
    <form @submit.prevent="handleSubmit">
      <div class="form-row">
        <div class="form-group">
          <label>Name <span class="hint-inline">Display name</span></label>
          <input type="text" v-model="form.name" placeholder="My Symfony App" required>
        </div>
        <div class="form-group">
          <label>Domain <span class="hint-inline">Add to hosts file</span></label>
          <input type="text" v-model="form.domain" placeholder="myapp.test" required>
        </div>
        <div class="form-group">
          <label>Target <span class="hint-inline">container:port</span></label>
          <input type="text" v-model="form.target" placeholder="myapp-nginx-1:80" required>
          <span class="field-hint">💡 Run <code>docker compose ps</code> to find container names</span>
        </div>
      </div>
      <div class="form-row" style="align-items: flex-end;">
        <div class="checkbox-group">
          <input type="checkbox" id="enabled" v-model="form.enabled">
          <label for="enabled" style="margin: 0;">Enable immediately</label>
        </div>
        <button type="submit" class="btn btn-primary">Add Route</button>
      </div>
    </form>
  </div>
</template>

<script>
export default {
  name: 'RouteForm',
  emits: ['submit'],
  data() {
    return {
      form: {
        name: '',
        domain: '',
        target: '',
        enabled: true,
      },
    }
  },
  methods: {
    handleSubmit() {
      this.$emit('submit', { ...this.form })
      this.form = { name: '', domain: '', target: '', enabled: true }
    },
  },
}
</script>
