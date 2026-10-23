<script setup>
import { ref } from 'vue'
import { Cancel01Icon, Download01Icon, UndoIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/vue'
import { yamlFilename } from '../lib/editor.js'
import ThemeStyleTab from './ThemeStyleTab.vue'

const props = defineProps({
  modelValue: { type: String, required: true },
  title: { type: String, default: '' },
  error: { type: String, default: '' },
  dirty: { type: Boolean, default: false },
  resolvedStyle: { type: Object, default: null }
})

const emit = defineEmits(['update:modelValue', 'reset', 'close'])

const iconStrokeWidth = 2.5

const tabs = [
  { id: 'yaml', label: 'YAML' },
  { id: 'theme', label: 'Theme and Style' }
]
const activeTab = ref('yaml')

function download () {
  const blob = new Blob([props.modelValue], { type: 'application/yaml' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = yamlFilename(props.title)
  link.click()
  URL.revokeObjectURL(link.href)
}
</script>

<template>
  <section id="editor-pane" aria-label="Org editor">
    <div class="editor-bar">
      <div class="editor-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :id="'editor-tab-' + tab.id"
          :key="tab.id"
          type="button"
          role="tab"
          :aria-selected="activeTab === tab.id"
          :class="{ active: activeTab === tab.id }"
          @click="activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>
      <button type="button" class="inline-icon neutral" title="Close editor" @click="emit('close')">
        <HugeiconsIcon
          :icon="Cancel01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Close</span>
      </button>
    </div>

    <textarea
      v-if="activeTab === 'yaml'"
      role="tabpanel"
      :value="modelValue"
      spellcheck="false"
      autocapitalize="off"
      autocomplete="off"
      aria-label="Org YAML"
      @input="emit('update:modelValue', $event.target.value)"
    />
    <ThemeStyleTab
      v-else
      role="tabpanel"
      :source="modelValue"
      :resolved="resolvedStyle"
      @change="emit('update:modelValue', $event)"
    />

    <p v-if="error" class="editor-status bad" role="status">{{ error }}</p>
    <p v-else class="editor-status good" role="status">Valid. The chart shows your changes.</p>

    <p class="editor-note">Chart YAML is a preview. Download it to keep chart edits. Shared themes are saved on the server.</p>

    <div class="editor-bar">
      <button type="button" class="inline-icon" :disabled="!dirty" @click="emit('reset')">
        <HugeiconsIcon
          :icon="UndoIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Reset</span>
      </button>
      <span class="editor-actions">
        <button type="button" class="inline-icon" title="Download the YAML" @click="download">
          <HugeiconsIcon
            :icon="Download01Icon"
            width="1em"
            height="1em"
            :strokeWidth="iconStrokeWidth"
            aria-hidden="true"
          />
          <span>Download</span>
        </button>
      </span>
    </div>
  </section>
</template>
