import { ref } from 'vue'

// Whether the YAML editor pane is open. Shared by the toolbar and the chart view.
export const editorOpen = ref(false)

export function debounce (fn, wait) {
  let timer = null

  const debounced = (...args) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), wait)
  }

  debounced.cancel = () => clearTimeout(timer)

  return debounced
}

// Wraps async calls so a slow response that arrives after a newer one can be
// recognised (current: false) and ignored. Errors are returned, not thrown,
// so a stale failure can be ignored too.
export function latestOnly () {
  let latest = 0

  return async (fn) => {
    const call = ++latest
    try {
      const value = await fn()
      return { current: call === latest, value }
    } catch (error) {
      return { current: call === latest, error }
    }
  }
}

// A file name for the chart, without an extension.
export function chartFilename (title) {
  const slug = (title || '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')

  return slug || 'org'
}

export function yamlFilename (title) {
  return chartFilename(title) + '.yml'
}
