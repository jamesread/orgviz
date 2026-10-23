import { describe, expect, it } from 'vitest'
import { personLevel, renderNodeContent } from './renderChart.js'

const allOn = { showAvatars: true, showJobTitles: true, showTeam: true, showExpandButtons: true }

function node (data, depth = 0) {
  return { depth, data: { avatarUrl: '/avatars/a.jpg', fullName: 'Ann', jobTitle: 'CEO', team: 'Leadership', ...data } }
}

describe('renderNodeContent', () => {
  it('escapes names, titles and teams', () => {
    const html = renderNodeContent(node({ fullName: '<img src=x onerror=alert(1)>', jobTitle: 'R&D "lead"', team: '<b>' }), allOn)

    expect(html).not.toContain('<img src=x')
    expect(html).toContain('&lt;img src=x onerror=alert(1)&gt;')
    expect(html).toContain('R&amp;D &quot;lead&quot;')
    expect(html).toContain('&lt;b&gt;')
  })

  it('falls back to the default avatar when the image is missing', () => {
    expect(renderNodeContent(node({}), allOn)).toContain("this.onerror=null;this.src='/avatars/default.png'")
  })

  it('shows everything that is switched on', () => {
    const html = renderNodeContent(node({}), allOn)

    expect(html).toContain('class="avatar"')
    expect(html).toContain('<p class="job-title">CEO</p>')
    expect(html).toContain('<p class="team">Leadership</p>')
  })

  it('leaves out what is switched off', () => {
    const html = renderNodeContent(node({}), { ...allOn, showAvatars: false, showJobTitles: false, showTeam: false })

    expect(html).not.toContain('class="avatar"')
    expect(html).not.toContain('job-title')
    expect(html).not.toContain('class="team"')
  })

  it('leaves out an empty team', () => {
    expect(renderNodeContent(node({ team: '' }), allOn)).not.toContain('class="team"')
  })

  it('marks the tree layer so a layer can share CSS', () => {
    expect(renderNodeContent(node({}), allOn)).toContain('class="person level-0"')
    expect(renderNodeContent(node({}, 2), allOn)).toContain('class="person level-2"')
  })

  it('marks the selected person for interactive editing', () => {
    const html = renderNodeContent(node({ id: 4 }), allOn, { selectedId: 4 })

    expect(html).toContain('data-person-id="4"')
    expect(html).toContain('person level-0 selected')
    expect(html).toContain('aria-current="true"')
    expect(renderNodeContent(node({ id: 4 }), allOn, { selectedId: 1 })).not.toContain(' selected')
  })
})

describe('personLevel', () => {
  it('uses the node depth and treats a missing one as the root', () => {
    expect(personLevel({ depth: 3 })).toBe(3)
    expect(personLevel({ depth: 1.8 })).toBe(1)
    expect(personLevel({})).toBe(0)
    expect(personLevel({ depth: -1 })).toBe(0)
  })
})
