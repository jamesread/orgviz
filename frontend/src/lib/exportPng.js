import { chartFilename } from './editor.js'

const SVG_NS = 'http://www.w3.org/2000/svg'
const CUSTOM_CSS_ID = 'chart-custom-css'
const CARD_SELECTOR = /person|avatar|node-button|job-title|\.team|path\.link|level-/
const EXPORT_TIMEOUT_MS = 30000
const IMPORT_RULE = /@import\s+[^;]+;/g

function readRules (sheet) {
  try {
    return [...sheet.cssRules]
  } catch {
    // Stylesheets from other origins cannot be read.
    return []
  }
}

export function stripImports (css) {
  return String(css || '').replace(IMPORT_RULE, '').trim()
}

export function readCustomCss (doc = document) {
  return doc.getElementById(CUSTOM_CSS_ID)?.textContent || ''
}

// The exported image is the chart's SVG on its own, without the page's
// stylesheets. Card rules come from the page; custom CSS is appended last
// from its source text so @import (which makes cssRules unreadable) cannot
// drop the editor styles.
export function collectChartCss (sheets, customCss = '') {
  const css = []
  let custom = customCss

  for (const sheet of sheets) {
    if (sheet.ownerNode?.id === CUSTOM_CSS_ID) {
      if (!custom) {
        custom = sheet.ownerNode.textContent || ''
      }
      continue
    }

    for (const rule of readRules(sheet)) {
      if (CARD_SELECTOR.test(rule.selectorText || '')) {
        css.push(rule.cssText)
      }
    }
  }

  const editorCss = stripImports(custom)
  if (editorCss) {
    css.push(editorCss)
  }

  return css.join('\n')
}

function pageTextCss () {
  const body = getComputedStyle(document.body)
  return `foreignObject { font-family: ${body.fontFamily}; color: ${body.color}; }`
}

function embedExportCss (svg, css) {
  const root = document.createElementNS(SVG_NS, 'style')
  root.textContent = css
  svg.insertBefore(root, svg.firstChild)

  const extras = []
  for (const div of svg.querySelectorAll('.node-foreign-object-div')) {
    const style = document.createElement('style')
    style.textContent = css
    div.insertBefore(style, div.firstChild)
    extras.push(style)
  }

  return () => {
    root.remove()
    for (const style of extras) {
      style.remove()
    }
  }
}

// Saves the whole chart as a PNG named after the chart title.
export function exportChartPng (chart, title, customCss = '') {
  const svg = chart.getChartState().svg.node()
  const hideButtons = svg.closest('.no-expand-buttons') ? '.node-button-g { display: none; }' : ''
  const css = [
    pageTextCss(),
    collectChartCss(document.styleSheets, customCss || readCustomCss()),
    hideButtons
  ].join('\n')
  const removeCss = embedExportCss(svg, css)

  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      removeCss()
      reject(new Error('The PNG export did not finish.'))
    }, EXPORT_TIMEOUT_MS)

    chart
      .imageName(chartFilename(title))
      .exportImg({
        full: true,
        scale: 2,
        // d3-org-chart defaults to #FAFAFA; keep the PNG background empty.
        backgroundColor: 'transparent',
        onLoad: () => {
          clearTimeout(timer)
          removeCss()
          resolve()
        }
      })
  })
}
