import { isMap, parseDocument } from 'yaml'

// The feature switches in an org file's `style.features` block. Defaults match
// orgs.StyleFeatures.Resolve on the server.
export const FEATURES = [
  { key: 'avatars', label: 'Avatars', default: true },
  { key: 'jobTitles', label: 'Job titles', default: true },
  { key: 'team', label: 'Team', default: false },
  { key: 'expandButtons', label: 'Expand/collapse buttons', default: true }
]

function defaultFeatures () {
  return {
    ...Object.fromEntries(FEATURES.map((feature) => [feature.key, feature.default])),
    compact: true
  }
}

// Reads the style block from org YAML, for the Theme and Style tab.
export function readStyle (source) {
  const doc = parseDocument(source)
  if (doc.errors.length > 0) {
    return { ok: false }
  }

  const features = defaultFeatures()
  for (const key of Object.keys(features)) {
    const value = doc.getIn(['style', 'features', key])
    if (typeof value === 'boolean') {
      features[key] = value
    }
  }

  const css = doc.getIn(['style', 'css'])
  const backgroundColor = doc.getIn(['style', 'backgroundColor'])
  const theme = doc.getIn(['style', 'theme'])

  return {
    ok: true,
    theme: typeof theme === 'string' ? theme : '',
    css: typeof css === 'string' ? css : '',
    backgroundColor: typeof backgroundColor === 'string' ? backgroundColor : '',
    features
  }
}

export const THEME_ID = /^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/

// Combines YAML style with the resolved style from the server, so the form
// shows theme values when the chart only names a theme.
export function displayStyle (style, resolved) {
  if (!style.ok) {
    return style
  }

  return {
    ok: true,
    theme: style.theme,
    css: resolved?.css ?? style.css,
    backgroundColor: resolved?.backgroundColor ?? style.backgroundColor,
    features: resolvedFeatures(style.features, resolved)
  }
}

function resolvedFeatures (features, resolved) {
  if (!resolved) {
    return features
  }

  return {
    avatars: resolved.showAvatars,
    jobTitles: resolved.showJobTitles,
    team: resolved.showTeam,
    expandButtons: resolved.showExpandButtons,
    compact: resolved.compact
  }
}

const HEX_6 = /^#[0-9a-f]{6}$/i
const HEX_3 = /^#([0-9a-f])([0-9a-f])([0-9a-f])$/i

// Converts a CSS color string into the #rrggbb format required by <input type="color">.
export function pickerColor (color) {
  const value = (color || '').trim()
  if (HEX_6.test(value)) {
    return value.toLowerCase()
  }

  const short = HEX_3.exec(value)
  if (short) {
    return `#${short[1]}${short[1]}${short[2]}${short[2]}${short[3]}${short[3]}`.toLowerCase()
  }

  return '#ffffff'
}

// Returns source with one change to its style block: { css },
// { backgroundColor }, { theme, reset } or { feature, value }. Comments and
// layout elsewhere in the file are kept.
export function writeStyle (source, change) {
  const doc = parseDocument(source)

  ensureMap(doc, ['style'])

  if ('theme' in change) {
    applyTheme(doc, change.theme, change.reset)
  } else if ('css' in change) {
    setCss(doc, change.css)
  } else if ('backgroundColor' in change) {
    setBackgroundColor(doc, change.backgroundColor)
  } else {
    ensureMap(doc, ['style', 'features'])
    doc.setIn(['style', 'features', change.feature], change.value)
  }

  removeIfEmpty(doc, ['style', 'features'])
  removeIfEmpty(doc, ['style'])

  return doc.toString()
}

function applyTheme (doc, theme, reset) {
  if (reset) {
    doc.deleteIn(['style', 'css'])
    doc.deleteIn(['style', 'backgroundColor'])
    doc.deleteIn(['style', 'features'])
  }

  const trimmed = (theme || '').trim()
  if (trimmed === '') {
    doc.deleteIn(['style', 'theme'])
    return
  }

  doc.setIn(['style', 'theme'], trimmed)
}

function setBackgroundColor (doc, color) {
  const trimmed = color.trim()
  if (trimmed === '') {
    doc.deleteIn(['style', 'backgroundColor'])
    return
  }

  doc.setIn(['style', 'backgroundColor'], trimmed)
}

function setCss (doc, css) {
  if (css === '') {
    doc.deleteIn(['style', 'css'])
    return
  }

  const node = doc.createNode(css)
  if (css.includes('\n')) {
    node.type = 'BLOCK_LITERAL'
  }
  doc.setIn(['style', 'css'], node)
}

function ensureMap (doc, path) {
  if (!isMap(doc.getIn(path, true))) {
    doc.setIn(path, doc.createNode({}))
  }
}

function removeIfEmpty (doc, path) {
  const node = doc.getIn(path, true)
  if (isMap(node) && node.items.length === 0) {
    doc.deleteIn(path)
  }
}
