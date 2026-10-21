<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Header from 'picocrank/vue/components/Header.vue'
import { client } from './lib/api.js'

const route = useRoute()
const router = useRouter()

const charts = ref([])
const version = ref('')
const isCompact = ref(true)

const layoutLabel = computed(() => isCompact.value ? 'Expand Layout' : 'Compact Layout')
const activeChartId = computed(() => route.params.chartId || '0')

onMounted(async () => {
  const settings = await client.getClientInitialSettings({})
  charts.value = settings.charts || []
  version.value = settings.version || ''
})

function openChart (chartId) {
  router.push({ name: 'chart', params: { chartId } })
}

function toggleLayout () {
  isCompact.value = !isCompact.value
  if (window.chart) {
    window.chart.compact(isCompact.value).render().fit()
  }
}
</script>

<template>
  <Header
    title="orgviz"
    logo-url="/logo.png"
    :sidebar-enabled="false"
  >
    <template #toolbar>
      <nav class="fg1">
        <ul id="charts-list">
          <li v-for="chart in charts" :key="chart.chartId">
            <a
              href="#"
              :class="{ active: chart.chartId === activeChartId }"
              @click.prevent="openChart(chart.chartId)"
            >{{ chart.title }}</a>
          </li>
        </ul>
      </nav>
      <nav>
        <ul>
          <li><a href="https://jamesread.github.io/orgviz/" target="_blank" rel="noreferrer">docs</a></li>
          <li><a href="https://github.com/jamesread/orgviz/" target="_blank" rel="noreferrer">GitHub repo</a></li>
        </ul>
      </nav>
      <button id="change-layout" type="button" class="neutral" @click="toggleLayout">
        {{ layoutLabel }}
      </button>
    </template>
  </Header>

  <div id="layout">
    <div id="content">
      <router-view />
      <footer>
        <span id="current-version">Version: {{ version || '?' }}</span>
      </footer>
    </div>
  </div>
</template>
