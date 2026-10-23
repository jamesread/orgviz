<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FitToScreenIcon, Image01Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/vue'
import EditorPane from '../components/EditorPane.vue'
import { client } from '../lib/api.js'
import { bindChartSelection, keepSelectedId } from '../lib/chartSelect.js'
import { debounce, editorOpen, latestOnly } from '../lib/editor.js'
import { exportChartPng } from '../lib/exportPng.js'
import { normalizePeople } from '../lib/normalizePeople.js'
import { DEFAULT_STYLE, applyCustomCss, createOrgChart } from '../lib/renderChart.js'

const iconStrokeWidth = 2.5

const route = useRoute()
const router = useRouter()

const title = ref('')
const originalSource = ref('')
const source = ref('')
const editorError = ref('')
const backgroundColor = ref('')
const chartStyle = ref(DEFAULT_STYLE)
const selectedPersonId = ref('')
const exporting = ref(false)
const exportError = ref('')
let unbindSelection = null

const dirty = computed(() => source.value !== originalSource.value)

// Chart loads and previews share one sequence, so a slow preview cannot
// overwrite a newer chart.
const latest = latestOnly()
const previewSource = debounce(runPreview, 300)

function renderPeople (people, style) {
  style = style || DEFAULT_STYLE
  chartStyle.value = style
  backgroundColor.value = style.backgroundColor || ''
  applyCustomCss(style.css)

  const chartEl = document.getElementById('chart')
  if (!chartEl) {
    return
  }

  chartEl.innerHTML = ''
  unbindSelection?.()

  window.people = normalizePeople(people || [])
  selectedPersonId.value = keepSelectedId(selectedPersonId.value, window.people)
  window.chart = createOrgChart('#chart', window.people, style, {
    getSelectedId: () => selectedPersonId.value
  })
  unbindSelection = bindChartSelection(chartEl, {
    getSelectedId: () => selectedPersonId.value,
    setSelectedId: (id) => { selectedPersonId.value = id }
  })
}

async function loadChart (chartId) {
  selectedPersonId.value = ''
  previewSource.cancel()

  const { current, value: data, error } = await latest(() => client.getChart({ chartId }))
  if (!current) {
    return
  }
  if (error) {
    throw error
  }

  // Old links use the chart's position; switch to its name.
  if (data.chartId !== chartId) {
    router.replace({ name: 'chart', params: { chartId: data.chartId } })
  }

  title.value = data.title
  originalSource.value = data.source
  source.value = data.source
  editorError.value = ''

  renderPeople(data.people, data.style)
}

async function runPreview (yaml) {
  const { current, value, error } = await latest(() => client.previewChart({ source: yaml }))
  if (!current) {
    return
  }

  if (error) {
    editorError.value = error.rawMessage || error.message
    return
  }

  editorError.value = ''
  renderPeople(value.people, value.style)
}

function onEdit (yaml) {
  source.value = yaml
  previewSource(yaml)
}

function recenter () {
  window.chart?.fit()
}

async function exportPng () {
  if (!window.chart) {
    return
  }

  exporting.value = true
  exportError.value = ''

  try {
    await exportChartPng(window.chart, title.value, chartStyle.value.css)
  } catch (err) {
    exportError.value = err.message
  } finally {
    exporting.value = false
  }
}

function resetSource () {
  onEdit(originalSource.value)
}

onMounted(() => {
  loadChart(route.params.chartId || '0')
})

watch(() => route.params.chartId, (chartId) => {
  if (chartId) {
    loadChart(chartId)
  }
})

watch(editorOpen, () => {
  // The chart area changes width when the pane opens or closes, and
  // d3-org-chart only measures its container when it renders.
  requestAnimationFrame(() => window.chart?.render().fit())
})

onBeforeUnmount(() => {
  unbindSelection?.()
  unbindSelection = null
  selectedPersonId.value = ''
  previewSource.cancel()
  backgroundColor.value = ''
  applyCustomCss('')
  window.chart = null
  window.people = null
})
</script>

<template>
  <main :class="{ editing: editorOpen }" :style="{ backgroundColor }">
    <div id="chart">Loading...</div>
    <div id="chart-controls">
      <button
        id="recenter-chart"
        type="button"
        class="inline-icon neutral"
        title="Fit the whole chart in view"
        @click="recenter"
      >
        <HugeiconsIcon
          :icon="FitToScreenIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Recenter</span>
      </button>
      <button
        id="export-png"
        type="button"
        class="inline-icon neutral"
        title="Save the whole chart as an image"
        :disabled="exporting"
        @click="exportPng"
      >
        <HugeiconsIcon
          :icon="Image01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>{{ exporting ? 'Exporting…' : 'Export PNG' }}</span>
      </button>
      <p v-if="exportError" class="chart-control-status" role="status">{{ exportError }}</p>
    </div>
    <EditorPane
      v-if="editorOpen"
      :model-value="source"
      :title="title"
      :error="editorError"
      :dirty="dirty"
      :resolved-style="chartStyle"
      @update:model-value="onEdit"
      @reset="resetSource"
      @close="editorOpen = false"
    />
  </main>
</template>
