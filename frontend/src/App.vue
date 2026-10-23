<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Cancel01Icon, HierarchySquare02Icon, SourceCodeIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/vue'
import Header from 'picocrank/vue/components/Header.vue'
import Sidebar from 'picocrank/vue/components/Sidebar.vue'
import { client } from './lib/api.js'
import { editorOpen } from './lib/editor.js'

const iconStrokeWidth = 2.5

const route = useRoute()
const sidebar = ref(null)

const charts = ref([])
const version = ref('')

const editorToggleIcon = computed(() => editorOpen.value ? Cancel01Icon : SourceCodeIcon)
const editorToggleLabel = computed(() => editorOpen.value ? 'Close Editor' : 'Edit YAML')
const activeChartId = computed(() => route.params.chartId || '0')

const navigation = computed(() => ({
  navigationLinks: [
    { type: 'section', name: 'organizations', title: 'Organizations' },
    ...charts.value.map((chart) => ({
      name: chart.chartId,
      title: chart.title,
      type: 'route',
      to: { name: 'chart', params: { chartId: chart.chartId } },
      icon: HierarchySquare02Icon
    }))
  ],
  isActive: (link) => link.name === activeChartId.value
}))

async function loadSettings () {
  const settings = await client.getClientInitialSettings({})
  charts.value = settings.charts || []
  version.value = settings.version || ''
}

onMounted(async () => {
  await loadSettings()
  if (localStorage.getItem('picocrank-sidebar-stuck') === null) {
    sidebar.value?.stick()
  }
})

// Charts can be added while the page is open, such as by an AI assistant
// through MCP, so refresh the list when opening one it does not have.
watch(activeChartId, (chartId) => {
  if (charts.value.length > 0 && !charts.value.some((chart) => chart.chartId === chartId)) {
    loadSettings()
  }
})

function toggleSidebar () {
  sidebar.value?.toggle()
}
</script>

<template>
  <Header
    title="orgviz"
    logo-url="/logo.png"
    :sidebar-enabled="true"
    :navigation="navigation"
    @toggleSidebar="toggleSidebar"
  >
    <template #toolbar>
      <button
        id="toggle-editor"
        type="button"
        class="inline-icon neutral"
        :aria-pressed="editorOpen"
        @click="editorOpen = !editorOpen"
      >
        <HugeiconsIcon
          :icon="editorToggleIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>{{ editorToggleLabel }}</span>
      </button>
    </template>
  </Header>

  <div id="layout">
    <Sidebar ref="sidebar" :navigation="navigation" />
    <div id="content">
      <router-view />
      <footer>
        <span id="current-version">Version: {{ version || '?' }}</span>
        <span>
          <a href="https://jamesread.github.io/orgviz/" target="_blank" rel="noreferrer">docs</a>
        </span>
        <span>
          <a href="https://github.com/jamesread/orgviz/" target="_blank" rel="noreferrer">GitHub repo</a>
        </span>
      </footer>
    </div>
  </div>
</template>
