<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { Cancel01Icon, Css3Icon, EraserIcon, FloppyDiskIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { client } from '../lib/api.js'
import { FEATURES, THEME_ID, displayStyle, pickerColor, readStyle, writeStyle } from '../lib/chartStyle.js'

const props = defineProps({
  source: { type: String, required: true },
  resolved: { type: Object, default: null }
})

const emit = defineEmits(['change'])

const iconStrokeWidth = 2.5

const layoutOptions = [
  { label: 'Compact', value: true },
  { label: 'Expanded', value: false }
]

const local = computed(() => readStyle(props.source))
const displayed = computed(() => displayStyle(local.value, props.resolved))

const themes = ref([])
const themeIdDraft = ref('')
const themeTitleDraft = ref('')
const saveError = ref('')
const saving = ref(false)

// Changes build on the newest YAML, including ones emitted before the parent
// has passed them back down, so quick successive changes are not lost.
let latestSource = props.source
watch(() => props.source, (source) => { latestSource = source })

function change (patch) {
  latestSource = writeStyle(latestSource, patch)
  emit('change', latestSource)
  return latestSource
}

const backgroundColorDraft = ref('')
let lastWrittenBackgroundColor = null

watch(() => displayed.value.backgroundColor, (backgroundColor) => {
  if (displayed.value.ok && backgroundColor !== lastWrittenBackgroundColor) {
    backgroundColorDraft.value = backgroundColor
  }
}, { immediate: true })

function onBackgroundColor (backgroundColor) {
  backgroundColorDraft.value = backgroundColor

  lastWrittenBackgroundColor = readStyle(change({ backgroundColor })).backgroundColor
}

const cssDraft = ref('')
let lastWrittenCss = null

watch(() => displayed.value.css, (css) => {
  if (displayed.value.ok && css !== lastWrittenCss) {
    cssDraft.value = css
  }
}, { immediate: true })

function onCss (css) {
  cssDraft.value = css

  lastWrittenCss = readStyle(change({ css })).css
}

// The CSS is edited in a dialog so the sidebar stays short.
const cssDialog = ref(null)

function openCssDialog () {
  cssDialog.value?.showModal()
}

function closeCssDialog () {
  cssDialog.value?.close()
}

function onFeature (feature, value) {
  change({ feature, value })
}

function onTheme (themeId) {
  change({ theme: themeId, reset: true })
}

onMounted(loadThemes)

watch(() => local.value.theme, (theme) => {
  if (theme) {
    themeIdDraft.value = theme
    themeTitleDraft.value = themeTitle(theme)
  }
}, { immediate: true })

async function loadThemes () {
  try {
    const data = await client.listThemes({})
    themes.value = data.themes || []
  } catch {
    themes.value = []
  }
}

function themeTitle (id) {
  return themes.value.find((theme) => theme.themeId === id)?.title || id
}

async function saveTheme () {
  if (!THEME_ID.test(themeIdDraft.value)) {
    saveError.value = 'Theme ID: letters, digits, - and _ (up to 64).'
    return
  }

  saving.value = true
  saveError.value = ''

  try {
    await client.saveTheme(themePayload())
    await loadThemes()
    change({ theme: themeIdDraft.value, reset: true })
  } catch (err) {
    saveError.value = err.rawMessage || err.message
  } finally {
    saving.value = false
  }
}

function themePayload () {
  const features = displayed.value.features

  return {
    themeId: themeIdDraft.value,
    title: themeTitleDraft.value || themeIdDraft.value,
    style: {
      css: cssDraft.value,
      backgroundColor: backgroundColorDraft.value,
      showAvatars: features.avatars,
      showJobTitles: features.jobTitles,
      showTeam: features.team,
      showExpandButtons: features.expandButtons,
      compact: features.compact
    }
  }
}
</script>

<template>
  <div class="theme-style-tab">
    <p v-if="!local.ok" class="editor-status bad" role="status">
      Fix the YAML on the YAML tab before changing the theme.
    </p>

    <fieldset class="theme-library" :disabled="!local.ok">
      <legend>Shared theme</legend>
      <label for="shared-theme">Use theme</label>
      <div class="theme-select">
        <select
          id="shared-theme"
          :value="local.ok ? local.theme : ''"
          @change="onTheme($event.target.value)"
        >
          <option value="">Custom</option>
          <option v-for="theme in themes" :key="theme.themeId" :value="theme.themeId">
            {{ theme.title }} ({{ theme.themeId }})
          </option>
        </select>
        <button
          type="button"
          class="inline-icon"
          aria-haspopup="dialog"
          :disabled="!local.ok"
          @click="openCssDialog"
        >
          <HugeiconsIcon
            :icon="Css3Icon"
            width="1em"
            height="1em"
            :strokeWidth="iconStrokeWidth"
            aria-hidden="true"
          />
          <span>Edit theme</span>
        </button>
      </div>

      <label for="save-theme-id">Save as</label>
      <div class="theme-save">
        <input
          id="save-theme-id"
          v-model="themeIdDraft"
          type="text"
          spellcheck="false"
          autocapitalize="off"
          autocomplete="off"
          placeholder="theme-id"
        >
        <input
          id="save-theme-title"
          v-model="themeTitleDraft"
          type="text"
          spellcheck="false"
          placeholder="Title"
        >
        <button type="button" class="inline-icon" :disabled="saving" @click="saveTheme">
          <HugeiconsIcon
            :icon="FloppyDiskIcon"
            width="1em"
            height="1em"
            :strokeWidth="iconStrokeWidth"
            aria-hidden="true"
          />
          <span>{{ saving ? 'Saving…' : 'Save theme' }}</span>
        </button>
      </div>
      <p v-if="saveError" class="editor-status bad" role="status">{{ saveError }}</p>
      <p class="editor-note">
        Themes are saved on the server in the <code>themes</code> directory and appear in this list for other charts.
      </p>
    </fieldset>

    <fieldset class="chart-features" :disabled="!local.ok">
      <legend>Features</legend>
      <div class="feature-checks">
        <label v-for="feature in FEATURES" :key="feature.key">
          <input
            type="checkbox"
            :checked="displayed.ok ? displayed.features[feature.key] : feature.default"
            @change="onFeature(feature.key, $event.target.checked)"
          >
          {{ feature.label }}
        </label>
      </div>
      <div class="layout-feature">
        <span id="chart-layout-label">Layout</span>
        <RadioGroup
          name="chart-layout"
          aria-labelledby="chart-layout-label"
          :model-value="displayed.ok ? displayed.features.compact : true"
          :options="layoutOptions"
          :disabled="!local.ok"
          @update:model-value="onFeature('compact', $event)"
        />
      </div>
    </fieldset>

    <label for="background-color">Background color</label>
    <div class="color-field">
      <input
        id="background-color-picker"
        type="color"
        :value="pickerColor(backgroundColorDraft)"
        :disabled="!local.ok"
        aria-label="Pick background color"
        @input="onBackgroundColor($event.target.value)"
      >
      <input
        id="background-color"
        type="text"
        :value="backgroundColorDraft"
        :disabled="!local.ok"
        spellcheck="false"
        autocapitalize="off"
        autocomplete="off"
        placeholder="#f8fafc or teal"
        @input="onBackgroundColor($event.target.value)"
      >
      <button
        v-if="backgroundColorDraft"
        type="button"
        class="inline-icon neutral"
        :disabled="!local.ok"
        @click="onBackgroundColor('')"
      >
        <HugeiconsIcon
          :icon="EraserIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Clear</span>
      </button>
    </div>

    <dialog
      ref="cssDialog"
      class="css-dialog"
      aria-labelledby="css-dialog-title"
      @click.self="closeCssDialog"
    >
      <div class="css-dialog-body">
        <div class="css-dialog-header">
          <h2 id="css-dialog-title">Custom CSS</h2>
          <button
            type="button"
            class="inline-icon neutral"
            aria-label="Close"
            @click="closeCssDialog"
          >
            <HugeiconsIcon
              :icon="Cancel01Icon"
              width="1em"
              height="1em"
              :strokeWidth="iconStrokeWidth"
              aria-hidden="true"
            />
          </button>
        </div>
        <textarea
          id="custom-css"
          aria-labelledby="css-dialog-title"
          :value="cssDraft"
          :disabled="!local.ok"
          spellcheck="false"
          autocapitalize="off"
          autocomplete="off"
          placeholder="div.person.level-0 { border-top-color: teal; }"
          @input="onCss($event.target.value)"
        />
        <p class="editor-note">
          Cards use <code>div.person</code>, <code>div.person.level-0</code> (root layer),
          <code>img.avatar</code>, <code>.person h3</code>,
          <code>.person .job-title</code> and <code>.person .team</code>.
          A shared theme is <code>style.theme</code>; other fields on this chart override it.
          Changes apply to the chart as you type.
        </p>
      </div>
    </dialog>
  </div>
</template>
