<template>
  <div class="container">
    <AppHeader
      :has-changes="hasUnappliedChanges"
      :reloading="reloading"
      @reload="reloadProxy"
      @export="exportConfig"
      @import="importConfig"
    />

    <!-- Navigation Tabs -->
    <div class="tabs">
      <button :class="['tab', { active: currentTab === 'routes' }]" @click="currentTab = 'routes'">
        Routes
      </button>
      <button :class="['tab', { active: currentTab === 'docs' }]" @click="currentTab = 'docs'">
        Documentation
      </button>
    </div>

    <!-- Routes Tab -->
    <div v-show="currentTab === 'routes'">
      <RouteForm @submit="handleAddRoute" />
      <RouteTable
        :routes="routes"
        :get-health-class="getHealthClass"
        :get-health-text="getHealthText"
        :get-health-tooltip="getHealthTooltip"
        :is-changed="isRouteChanged"
        @toggle="toggleRoute"
        @edit="openEditModal"
        @delete="handleDeleteRoute"
        @show-health="openHealthModal"
        @show-docs="currentTab = 'docs'"
      />
    </div>

    <!-- Documentation Tab -->
    <DocsTab v-show="currentTab === 'docs'" />

    <!-- Modals -->
    <EditRouteModal
      v-if="editingRoute"
      :route="editingRoute"
      @close="editingRoute = null"
      @save="handleSaveRoute"
    />

    <HealthModal
      v-if="healthDetails"
      :details="healthDetails"
      @close="healthDetails = null"
    />

    <LoadingModal
      v-if="reloading"
      message="Reloading proxy configuration..."
    />

    <ToastNotification
      :show="toast.show"
      :message="toast.message"
      :type="toast.type"
    />
  </div>
</template>

<script>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoutes, useHealth, useProxy, useToast } from './composables'

// Components
import AppHeader from './components/AppHeader.vue'
import RouteForm from './components/RouteForm.vue'
import RouteTable from './components/RouteTable.vue'
import DocsTab from './components/DocsTab.vue'
import ToastNotification from './components/ToastNotification.vue'
import EditRouteModal from './components/modals/EditRouteModal.vue'
import HealthModal from './components/modals/HealthModal.vue'
import LoadingModal from './components/modals/LoadingModal.vue'

export default {
  name: 'App',
  components: {
    AppHeader,
    RouteForm,
    RouteTable,
    DocsTab,
    ToastNotification,
    EditRouteModal,
    HealthModal,
    LoadingModal,
  },
  setup() {
    const { routes, fetchRoutes, createRoute, updateRoute, deleteRoute, toggleRoute } = useRoutes()
    const { getHealthClass, getHealthText, getHealthTooltip, getHealthDetails, startPolling, stopPolling } = useHealth()
    const { reloading, hasUnappliedChanges, isRouteChanged, fetchAppliedState, reloadProxy, exportConfig, importConfig } = useProxy()
    const { toast } = useToast()

    onMounted(() => {
      fetchRoutes()
      fetchAppliedState()
      startPolling()
    })

    onUnmounted(() => {
      stopPolling()
    })

    return {
      // State
      routes,
      reloading,
      hasUnappliedChanges,
      toast,

      // Methods
      fetchRoutes,
      createRoute,
      updateRoute,
      deleteRoute,
      toggleRoute,
      getHealthClass,
      getHealthText,
      getHealthTooltip,
      getHealthDetails,
      isRouteChanged,
      reloadProxy,
      exportConfig,
      importConfig,
    }
  },
  data() {
    return {
      currentTab: 'routes',
      editingRoute: null,
      healthDetails: null,
    }
  },
  methods: {
    async handleAddRoute(route) {
      await this.createRoute(route)
    },
    async handleSaveRoute(route) {
      await this.updateRoute(route.id, route)
      this.editingRoute = null
    },
    async handleDeleteRoute(route) {
      await this.deleteRoute(route.id, route.name)
    },
    openEditModal(route) {
      this.editingRoute = { ...route }
    },
    openHealthModal(id) {
      const details = this.getHealthDetails(id)
      if (details) {
        this.healthDetails = details
      }
    },
  },
}
</script>
