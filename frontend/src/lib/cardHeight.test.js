import { describe, expect, it } from 'vitest'
import { cardHeights, cardId, cardRows, equalizeRowHeights, estimatePersonHeight, personDepths } from './cardHeight.js'

const alice = { id: 1, parentId: null, fullName: 'Alice', jobTitle: 'CEO', team: 'Leadership' }
const bob = { id: 2, parentId: 1, fullName: 'Bob', jobTitle: 'VP', team: 'Product' }
const carol = { id: 3, parentId: 1, fullName: 'Carol', jobTitle: 'Engineer', team: 'Platform' }
const dan = { id: 4, parentId: 1, fullName: 'Dan', jobTitle: 'Designer', team: 'Design' }
const eve = { id: 5, parentId: 2, fullName: 'Eve', jobTitle: 'Intern', team: 'Product' }

const people = [alice, bob, carol, dan, eve]
const style = { showAvatars: true, showJobTitles: true, showTeam: false }

describe('personDepths', () => {
  it('counts steps from the root', () => {
    const depths = personDepths(people)

    expect(depths.get(1)).toBe(0)
    expect(depths.get(2)).toBe(1)
    expect(depths.get(3)).toBe(1)
    expect(depths.get(5)).toBe(2)
  })
})

describe('cardRows', () => {
  it('puts everyone at the same depth in one row when expanded', () => {
    const rows = cardRows(people, false)

    expect(rows.get('1')).toBe('d:0')
    expect(rows.get('2')).toBe('d:1')
    expect(rows.get('3')).toBe('d:1')
    expect(rows.get('4')).toBe('d:1')
    expect(rows.get('5')).toBe('d:2')
  })

  it('pairs compact leaf siblings and leaves managers on their depth row', () => {
    const rows = cardRows(people, true)

    expect(rows.get('1')).toBe('d:0')
    expect(rows.get('2')).toBe('d:1')
    expect(rows.get('3')).toBe('c:1:0')
    expect(rows.get('4')).toBe('c:1:0')
    expect(rows.get('5')).toBe('d:2')
  })
})

describe('equalizeRowHeights', () => {
  it('uses the tallest card in each row', () => {
    const rows = new Map([[1, 'a'], [2, 'a'], [3, 'b']])
    const natural = new Map([[1, 50], [2, 90], [3, 40]])

    expect(Object.fromEntries(equalizeRowHeights(natural, rows)))
      .toEqual({ 1: 90, 2: 90, 3: 40 })
  })

  it('matches numeric row ids to string measure keys', () => {
    const rows = new Map([[1, 'a'], [2, 'a']])
    const natural = new Map([['1', 40], ['2', 70]])

    expect(Object.fromEntries(equalizeRowHeights(natural, rows)))
      .toEqual({ 1: 70, 2: 70 })
  })
})

describe('estimatePersonHeight', () => {
  it('grows with wrapping titles and optional team', () => {
    const noAvatar = { ...style, showAvatars: false }
    const short = estimatePersonHeight({ fullName: 'Ann', jobTitle: 'CEO' }, noAvatar)
    const wrapped = estimatePersonHeight({
      fullName: 'Ann',
      jobTitle: 'Vice President of Platform Engineering and Developer Experience'
    }, noAvatar)
    const withTeam = estimatePersonHeight({ fullName: 'Ann', jobTitle: 'CEO', team: 'Leadership' }, { ...noAvatar, showTeam: true })

    expect(wrapped).toBeGreaterThan(short)
    expect(withTeam).toBeGreaterThan(short)
  })
})

describe('cardHeights', () => {
  it('makes compact-row cards equal and sized to the taller one', () => {
    const tall = { ...carol, jobTitle: 'Principal Engineer for Cloud Infrastructure Platforms' }
    const heights = cardHeights([alice, bob, tall, dan, eve], style, true)

    expect(heights.get('3')).toBe(heights.get('4'))
    expect(heights.get('3')).toBeGreaterThan(heights.get('2'))
  })

  it('makes an expanded depth share one height', () => {
    const tall = { ...carol, jobTitle: 'Principal Engineer for Cloud Infrastructure Platforms' }
    const heights = cardHeights([alice, bob, tall, dan, eve], style, false)

    expect(heights.get('2')).toBe(heights.get('3'))
    expect(heights.get('3')).toBe(heights.get('4'))
    expect(heights.get('3')).toBeGreaterThan(heights.get('1'))
  })
})

describe('cardId', () => {
  it('stringifies proto and hierarchy ids the same way', () => {
    expect(cardId({ id: 8 })).toBe('8')
    expect(cardId({ data: { id: 8 } })).toBe('8')
  })
})
