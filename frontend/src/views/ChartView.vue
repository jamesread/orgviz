<script setup>
import { onBeforeUnmount, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { client } from '../lib/api.js'
import { normalizePeople } from '../lib/normalizePeople.js'
import { createOrgChart } from '../lib/renderChart.js'

const route = useRoute()

async function loadChart (chartId) {
  const chartEl = document.getElementById('chart')
  if (!chartEl) {
    return
  }

  chartEl.innerHTML = ''

  const data = await client.getChart({ chartId })
  const people = normalizePeople(data.people || [])

  window.people = people
  window.isCompact = window.isCompact !== false
  window.chart = createOrgChart('#chart', people, window.isCompact)
}

onMounted(() => {
  loadChart(route.params.chartId || '0')
})

watch(() => route.params.chartId, (chartId) => {
  if (chartId) {
    loadChart(chartId)
  }
})

onBeforeUnmount(() => {
  window.chart = null
  window.people = null
})
</script>

<template>
  <main>
    <div id="chart">Loading...</div>
  </main>
</template>
