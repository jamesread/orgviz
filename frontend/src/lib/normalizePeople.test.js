import { describe, expect, it } from 'vitest'
import { normalizePeople } from './normalizePeople.js'

describe('normalizePeople', () => {
  it('makes the first person the root and fills missing avatars', () => {
    const people = normalizePeople([
      { id: 1, parentId: 0, fullName: 'Alice', avatarUrl: '' },
      { id: 2, parentId: 0, fullName: 'Bob', avatarUrl: '/avatars/bob.jpg' }
    ])

    expect(people[0].parentId).toBeNull()
    expect(people[0].avatarUrl).toBe('/avatars/default.png')
    expect(people[1].parentId).toBe(1)
    expect(people[1].avatarUrl).toBe('/avatars/bob.jpg')
  })
})
