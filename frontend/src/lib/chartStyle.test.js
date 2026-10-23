import { describe, expect, it } from 'vitest'
import { FEATURES, displayStyle, pickerColor, readStyle, writeStyle } from './chartStyle.js'

const base = `# Our org
title: Demo
people:
  - name: Alice # the boss
`

describe('readStyle', () => {
  it('uses the defaults when there is no style block', () => {
    expect(readStyle(base)).toEqual({
      ok: true,
      theme: '',
      css: '',
      backgroundColor: '',
      features: { avatars: true, jobTitles: true, team: false, expandButtons: true, compact: true }
    })
  })

  it('reads backgroundColor, css and features', () => {
    const style = readStyle(base + 'style:\n  backgroundColor: "#f8fafc"\n  css: "h3 { color: red; }"\n  features:\n    avatars: false\n    team: true\n')

    expect(style.backgroundColor).toBe('#f8fafc')
    expect(style.css).toBe('h3 { color: red; }')
    expect(style.features).toEqual({ avatars: false, jobTitles: true, team: true, expandButtons: true, compact: true })
  })

  it('reports YAML that does not parse', () => {
    expect(readStyle('people: [').ok).toBe(false)
  })
})

describe('writeStyle', () => {
  it('adds a feature switch and keeps comments', () => {
    const source = writeStyle(base, { feature: 'team', value: true })

    expect(source).toContain('# Our org')
    expect(source).toContain('# the boss')
    expect(readStyle(source).features.team).toBe(true)
  })

  it('writes a theme and clears local overrides', () => {
    let source = writeStyle(base, { css: 'h3 { color: red; }' })
    source = writeStyle(source, { theme: 'ocean', reset: true })

    expect(readStyle(source).theme).toBe('ocean')
    expect(readStyle(source).css).toBe('')
    expect(source).not.toContain('css:')

    source = writeStyle(source, { theme: '', reset: true })
    expect(readStyle(source).theme).toBe('')
    expect(source).not.toContain('style')
  })

  it('writes and clears backgroundColor', () => {
    let source = writeStyle(base, { backgroundColor: '#f8fafc' })
    expect(readStyle(source).backgroundColor).toBe('#f8fafc')

    source = writeStyle(source, { backgroundColor: '  ' })
    expect(readStyle(source).backgroundColor).toBe('')
    expect(source).not.toContain('style')
  })

  it('writes multi-line css as a block literal', () => {
    const css = 'div.person {\n  border-top-color: teal;\n}\n'
    const source = writeStyle(base, { css })

    expect(source).toContain('css: |')
    expect(readStyle(source).css).toBe(css)
  })

  it('removes the style block once it is empty', () => {
    let source = writeStyle(base, { css: 'h3 { color: red; }' })
    source = writeStyle(source, { css: '' })

    expect(source).not.toContain('style')
  })

  it('replaces an empty style key', () => {
    const source = writeStyle(base + 'style:\n', { feature: 'avatars', value: false })

    expect(readStyle(source).features.avatars).toBe(false)
  })

  it('writes compact layout', () => {
    const source = writeStyle(base, { feature: 'compact', value: false })

    expect(readStyle(source).features.compact).toBe(false)
  })

  it('has a label for every feature', () => {
    expect(FEATURES.map((feature) => feature.key)).toEqual(['avatars', 'jobTitles', 'team', 'expandButtons'])
    expect(FEATURES.every((feature) => feature.label)).toBe(true)
  })
})

describe('displayStyle', () => {
  it('uses resolved theme values when the YAML only names a theme', () => {
    const local = { ok: true, theme: 'ocean', css: '', backgroundColor: '', features: { avatars: true, jobTitles: true, team: false, expandButtons: true, compact: true } }
    const shown = displayStyle(local, { css: 'h3 { color: red; }', backgroundColor: '#eee', showAvatars: true, showJobTitles: true, showTeam: true, showExpandButtons: false, compact: false })

    expect(shown.css).toBe('h3 { color: red; }')
    expect(shown.backgroundColor).toBe('#eee')
    expect(shown.features.team).toBe(true)
    expect(shown.features.compact).toBe(false)
    expect(shown.theme).toBe('ocean')
  })
})

describe('pickerColor', () => {
  it('normalises hex colors and falls back to white', () => {
    expect(pickerColor('#F8FAFC')).toBe('#f8fafc')
    expect(pickerColor('#abc')).toBe('#aabbcc')
    expect(pickerColor('teal')).toBe('#ffffff')
    expect(pickerColor('')).toBe('#ffffff')
  })
})
