import { describe, expect, it, vi } from 'vitest'
import { applyNodeSelection, keepSelectedId, nextSelectedId } from './chartSelect.js'

describe('nextSelectedId', () => {
  it('selects a different card and clears the same one', () => {
    expect(nextSelectedId('', '3')).toBe('3')
    expect(nextSelectedId('3', '3')).toBe('')
    expect(nextSelectedId('3', '8')).toBe('8')
  })
})

describe('keepSelectedId', () => {
  it('drops a selection when that person is gone', () => {
    const people = [{ id: 1 }, { id: 2 }]

    expect(keepSelectedId('2', people)).toBe('2')
    expect(keepSelectedId('9', people)).toBe('')
    expect(keepSelectedId('', people)).toBe('')
  })
})

describe('applyNodeSelection', () => {
  it('marks only the matching card', () => {
    const cards = [
      { dataset: { personId: '1' }, classList: { toggle: vi.fn() }, setAttribute: vi.fn(), removeAttribute: vi.fn(), closest: () => null },
      { dataset: { personId: '2' }, classList: { toggle: vi.fn() }, setAttribute: vi.fn(), removeAttribute: vi.fn(), closest: () => null }
    ]

    applyNodeSelection({ querySelectorAll: () => cards }, '2')

    expect(cards[0].classList.toggle).toHaveBeenCalledWith('selected', false)
    expect(cards[1].classList.toggle).toHaveBeenCalledWith('selected', true)
    expect(cards[1].setAttribute).toHaveBeenCalledWith('aria-current', 'true')
  })
})
