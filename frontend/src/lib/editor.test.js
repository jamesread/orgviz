import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { chartFilename, debounce, latestOnly, yamlFilename } from './editor.js'

describe('debounce', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('calls once with the last arguments after the wait', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 300)

    debounced('a')
    debounced('b')
    vi.advanceTimersByTime(299)
    expect(fn).not.toHaveBeenCalled()

    vi.advanceTimersByTime(1)
    expect(fn).toHaveBeenCalledTimes(1)
    expect(fn).toHaveBeenCalledWith('b')
  })

  it('can be cancelled', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 300)

    debounced('a')
    debounced.cancel()
    vi.advanceTimersByTime(300)
    expect(fn).not.toHaveBeenCalled()
  })
})

describe('latestOnly', () => {
  it('reports only the most recent call as current', async () => {
    let resolveFirst
    const first = new Promise((resolve) => { resolveFirst = resolve })
    const run = latestOnly()

    const slow = run(() => first)
    const fast = run(() => Promise.resolve('second'))

    expect(await fast).toEqual({ current: true, value: 'second' })
    resolveFirst('first')
    expect(await slow).toEqual({ current: false, value: 'first' })
  })

  it('returns errors instead of throwing them', async () => {
    const run = latestOnly()
    const failure = new Error('bad yaml')

    expect(await run(() => Promise.reject(failure))).toEqual({ current: true, error: failure })
  })
})

describe('yamlFilename', () => {
  it('slugs the chart title', () => {
    expect(yamlFilename('Example Company')).toBe('example-company.yml')
    expect(yamlFilename('  R&D / Ops ')).toBe('r-d-ops.yml')
  })

  it('falls back when the title is empty', () => {
    expect(yamlFilename('')).toBe('org.yml')
    expect(yamlFilename('!!!')).toBe('org.yml')
  })
})

describe('chartFilename', () => {
  it('slugs the title and adds no extension', () => {
    expect(chartFilename('Example Company')).toBe('example-company')
    expect(chartFilename('')).toBe('org')
  })
})
