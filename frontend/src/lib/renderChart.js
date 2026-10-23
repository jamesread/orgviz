import * as d3 from 'd3'
import { OrgChart } from 'd3-org-chart'
import { CARD_WIDTH, cardHeights, cardId, contentBoxHeight, estimatePersonHeight, heightFor, personDepths } from './cardHeight.js'
import { stripImports } from './exportPng.js'

const DEFAULT_AVATAR = '/avatars/default.png'
const CUSTOM_CSS_ID = 'chart-custom-css'

// Used when the server sends no style, matching orgs.StyleFeatures.Resolve.
export const DEFAULT_STYLE = {
  css: '',
  backgroundColor: '',
  theme: '',
  showAvatars: true,
  showJobTitles: true,
  showTeam: false,
  showExpandButtons: true,
  compact: true
}

function escapeHtml (value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function renderAvatar (d, name) {
  return `<img src="${escapeHtml(d.data.avatarUrl)}" alt="${name}" class="avatar"
              onerror="this.onerror=null;this.src='${DEFAULT_AVATAR}'" />`
}

// Tree depth from d3-org-chart: 0 is the root, 1 is their reports, and so on.
export function personLevel (d) {
  const depth = Number(d?.depth)
  if (!Number.isFinite(depth) || depth < 0) {
    return 0
  }

  return Math.floor(depth)
}

// Names, titles and teams can come from the YAML editor, so they are escaped.
// Avatars are guessed from first names, so a missing image falls back to the
// default.
export function renderNodeContent (d, style = DEFAULT_STYLE, options = {}) {
  const name = escapeHtml(d.data.fullName)
  const id = cardId(d)
  const selected = id !== '' && String(options.selectedId ?? '') === id
  const avatar = style.showAvatars ? renderAvatar(d, name) : ''
  const jobTitle = style.showJobTitles ? `<p class="job-title">${escapeHtml(d.data.jobTitle)}</p>` : ''
  const team = style.showTeam && d.data.team ? `<p class="team">${escapeHtml(d.data.team)}</p>` : ''

  return `<div class="person level-${personLevel(d)}${selected ? ' selected' : ''}" data-person-id="${escapeHtml(id)}"${selected ? ' aria-current="true"' : ''}>
            ${avatar}
            <div class="person-info">
            <h3>${name}</h3>
            ${jobTitle}
            ${team}
            </div>
          </div>`
}

// Custom CSS comes from the chart's style block. textContent keeps it as CSS.
export function applyCustomCss (css) {
  let el = document.getElementById(CUSTOM_CSS_ID)
  if (!el) {
    el = document.createElement('style')
    el.id = CUSTOM_CSS_ID
    document.head.appendChild(el)
  }

  el.textContent = css || ''
}

function measureCards (people, style, width) {
  const depths = personDepths(people)
  const host = document.createElement('div')
  host.setAttribute('aria-hidden', 'true')
  host.setAttribute('data-card-measure', '')
  host.style.cssText = `position:absolute;left:-9999px;top:0;width:${width}px;font:12px sans-serif;pointer-events:none;`

  const css = document.createElement('style')
  const themeCss = stripImports(style.css).replace(/#chart\b/g, '[data-card-measure]')
  css.textContent = `[data-card-measure] .person { height: auto !important; }\n${themeCss}`
  host.appendChild(css)

  const nodes = people.map((person) => {
    const wrap = document.createElement('div')
    wrap.innerHTML = renderNodeContent({ data: person, depth: depths.get(person.id) }, style)
    host.appendChild(wrap)
    return { id: person.id, el: wrap.firstElementChild }
  })

  document.body.appendChild(host)
  const heights = new Map()
  for (const { id, el } of nodes) {
    heights.set(String(id), contentBoxHeight(el))
  }
  host.remove()
  return heights
}

export function createOrgChart (container, people, style = DEFAULT_STYLE, options = {}) {
  document.querySelector(container)?.classList.toggle('no-expand-buttons', !style.showExpandButtons)

  const compact = style.compact !== false
  const canMeasure = typeof document !== 'undefined' && document.body
  const heights = cardHeights(people, style, compact, CARD_WIDTH, canMeasure ? measureCards : undefined)

  const chart = new OrgChart()
    .nodeWidth(() => CARD_WIDTH)
    .nodeHeight((d) => heightFor(heights, d) ?? heightFor(heights, d.data) ?? estimatePersonHeight(d.data || {}, style))
    .container(container)
    .childrenMargin(() => 50)
    .compactMarginBetween(() => 35)
    .compactMarginPair(() => 30)
    .neighbourMargin(() => 20)
    .compact(compact)
    .linkUpdate(function () {
      return d3.select(this).attr('stroke', '#666')
    })
    .nodeContent((d) => renderNodeContent(d, style, { selectedId: options.getSelectedId?.() ?? options.selectedId }))

  chart.data(people).render()

  // Without the buttons nobody could open a collapsed branch.
  if (!style.showExpandButtons) {
    chart.expandAll().fit()
  }

  return chart
}
