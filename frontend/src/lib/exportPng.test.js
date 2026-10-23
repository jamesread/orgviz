import { describe, expect, it } from 'vitest'
import { collectChartCss, stripImports } from './exportPng.js'

function sheet (rules, id = '', textContent = '') {
  return { ownerNode: { id, textContent }, cssRules: rules }
}

describe('collectChartCss', () => {
  it('keeps card rules and drops page rules', () => {
    const css = collectChartCss([
      sheet([
        { selectorText: 'div.person', cssText: 'div.person { border: 1px solid; }' },
        { selectorText: 'img.avatar', cssText: 'img.avatar { width: 80px; }' },
        { selectorText: '#chart.no-expand-buttons .node-button-g', cssText: '#chart.no-expand-buttons .node-button-g { display: none; }' },
        { selectorText: 'body', cssText: 'body { margin: 0; }' }
      ])
    ])

    expect(css).toContain('div.person { border: 1px solid; }')
    expect(css).toContain('img.avatar { width: 80px; }')
    expect(css).toContain('.node-button-g')
    expect(css).not.toContain('body')
  })

  it('appends editor css after card rules so it overrides defaults', () => {
    const css = collectChartCss([
      sheet([{ selectorText: 'div.person', cssText: 'div.person { border-top-color: purple; }' }]),
      sheet([], 'chart-custom-css', 'div.person { border-top-color: teal; }')
    ])

    expect(css).toContain('div.person { border-top-color: teal; }')
    expect(css.indexOf('purple')).toBeLessThan(css.indexOf('teal'))
  })

  it('uses the provided custom css when the stylesheet cannot be read', () => {
    const unreadable = {
      ownerNode: { id: 'chart-custom-css', textContent: '' },
      get cssRules () { throw new Error('cross-origin') }
    }

    const css = collectChartCss(
      [unreadable],
      "@import url('https://fonts.googleapis.com/css2?family=x');\ndiv.person { border-top-color: teal; }"
    )

    expect(css).toContain('div.person { border-top-color: teal; }')
    expect(css).not.toContain('@import')
  })

  it('skips sheets it cannot read', () => {
    const unreadable = { ownerNode: { id: '' }, get cssRules () { throw new Error('cross-origin') } }

    expect(collectChartCss([unreadable])).toBe('')
  })
})

describe('stripImports', () => {
  it('removes @import rules and keeps the rest', () => {
    expect(stripImports("@import url('https://example.com/x.css');\n\ndiv.person { color: red; }"))
      .toBe('div.person { color: red; }')
  })
})
