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
      <div class="tabs-left">
        <button :class="['tab', { active: currentTab === 'routes' }]" @click="currentTab = 'routes'">
          Routes
        </button>
        <button :class="['tab', { active: currentTab === 'updates' }]" @click="currentTab = 'updates'">
          Updates
        </button>
        <button :class="['tab', { active: currentTab === 'agent' }]" @click="currentTab = 'agent'">
          Host Agent
        </button>
        <button :class="['tab', { active: currentTab === 'docs' }]" @click="currentTab = 'docs'">
          Documentation
        </button>
      </div>
      <button
        class="agent-config-btn"
        :disabled="!agentReachable"
        :title="agentReachable ? 'Open Agent Config Panel' : 'Agent is not running'"
        @click="openAgentConfig"
      >
        ⚙️ Agent Config
      </button>
    </div>

    <!-- Routes Tab -->
    <div v-show="currentTab === 'routes'">
      <UpdateBanner :updateInfo="updateInfo" />
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

    <!-- Updates Tab -->
    <UpdateTab v-show="currentTab === 'updates'" @switch-tab="handleSwitchTab" @show-toast="handleToast" />

    <!-- Host Agent Tab -->
    <AgentTab v-show="currentTab === 'agent'" />

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
import AgentTab from './components/AgentTab.vue'
import UpdateTab from './components/UpdateTab.vue'
import ToastNotification from './components/ToastNotification.vue'
import UpdateBanner from './components/UpdateBanner.vue'
import EditRouteModal from './components/modals/EditRouteModal.vue'
import HealthModal from './components/modals/HealthModal.vue'
import LoadingModal from './components/modals/LoadingModal.vue'
import { agentApi } from './api'

export default {
  name: 'App',
  components: {
    AppHeader,
    RouteForm,
    RouteTable,
    DocsTab,
    AgentTab,
    UpdateTab,
    ToastNotification,
    UpdateBanner,
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
      agentReachable: false,
      updateInfo: null,
    }
  },
  mounted() {
    this.checkAgent()
    this.checkForUpdates()
    this._agentInterval = setInterval(() => this.checkAgent(), 5000)
    // Check for updates every 30 minutes
    this._updateInterval = setInterval(() => this.checkForUpdates(), 30 * 60 * 1000)
  },
  beforeUnmount() {
    if (this._agentInterval) clearInterval(this._agentInterval)
    if (this._updateInterval) clearInterval(this._updateInterval)
  },
  methods: {
    async checkAgent() {
      try {
        const controller = new AbortController()
        const timeout = setTimeout(() => controller.abort(), 2000)
        const resp = await fetch('http://localhost:9099/api/status', { signal: controller.signal })
        clearTimeout(timeout)
        this.agentReachable = resp.ok
      } catch {
        this.agentReachable = false
      }
    },
    openAgentConfig() {
      if (this.agentReachable) {
        window.location.href = 'http://localhost:9099'
      }
    },
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
    async checkForUpdates() {
      try {
        const data = await agentApi.checkUpdates()
        this.updateInfo = data
      } catch (error) {
        // Silently fail - agent might not be running
        console.debug('Failed to check for updates:', error.message)
      }
    },
    handleSwitchTab(tab) {
      this.currentTab = tab
    },
    handleToast({ message, type }) {
      this.toast.message = message
      this.toast.type = type
      this.toast.show = true
      setTimeout(() => {
        this.toast.show = false
      }, 3000)
    },
  },
}
</script>
