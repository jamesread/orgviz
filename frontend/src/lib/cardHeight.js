export const CARD_WIDTH = 250

const PAD_X = 20
const PAD_Y = 24
const AVATAR = 80
const AVATAR_GAP = 16
const CHAR_WIDTH = 7
const NAME_LINE = 22
const TITLE_LINE = 18
const TEAM_LINE = 16

export function personDepths (people) {
  const byId = new Map(people.map((person) => [person.id, person]))
  const depths = new Map()

  function depthOf (id) {
    if (depths.has(id)) {
      return depths.get(id)
    }

    depths.set(id, 0)
    const person = byId.get(id)
    if (!person || person.parentId == null || !byId.has(person.parentId)) {
      return 0
    }

    const depth = depthOf(person.parentId) + 1
    depths.set(id, depth)
    return depth
  }

  for (const person of people) {
    depthOf(person.id)
  }

  return depths
}

function childrenByParent (people) {
  const childrenOf = new Map()

  for (const person of people) {
    if (person.parentId == null) {
      continue
    }

    const kids = childrenOf.get(person.parentId) || []
    kids.push(person)
    childrenOf.set(person.parentId, kids)
  }

  return childrenOf
}

export function cardRows (people, compact) {
  const depths = personDepths(people)
  const rows = new Map()

  for (const person of people) {
    rows.set(cardId(person), `d:${depths.get(person.id)}`)
  }

  if (!compact) {
    return rows
  }

  const childrenOf = childrenByParent(people)

  for (const [parentId, kids] of childrenOf) {
    const leaves = kids.filter((kid) => !childrenOf.has(kid.id))
    if (leaves.length < 2) {
      continue
    }

    leaves.forEach((person, i) => {
      rows.set(cardId(person), `c:${parentId}:${Math.floor(i / 2)}`)
    })
  }

  return rows
}

export function cardId (person) {
  const id = person?.id ?? person?.data?.id
  return id == null ? '' : String(id)
}

export function heightFor (heights, person) {
  const id = cardId(person)
  return heights.get(id) ?? heights.get(person?.id) ?? heights.get(Number(id))
}

// Content height only: theme `height: 100%` would otherwise report the
// stretched card, not the text and avatar.
export function contentBoxHeight (el) {
  if (!el) {
    return 0
  }

  const cs = getComputedStyle(el)
  const pad = Number.parseFloat(cs.paddingTop) + Number.parseFloat(cs.paddingBottom)
  const border = Number.parseFloat(cs.borderTopWidth) + Number.parseFloat(cs.borderBottomWidth)
  let body = 0
  for (const child of el.children) {
    body = Math.max(body, child.offsetHeight)
  }

  return Math.ceil(body + pad + border)
}

export function equalizeRowHeights (naturalById, rows) {
  const maxByRow = new Map()

  for (const [id, row] of rows) {
    const height = heightFor(naturalById, { id }) || 0
    maxByRow.set(row, Math.max(maxByRow.get(row) || 0, height))
  }

  const heights = new Map()
  for (const [id, row] of rows) {
    heights.set(String(id), maxByRow.get(row))
  }

  return heights
}

function wrapCount (text, maxChars) {
  const words = String(text || '').trim().split(/\s+/).filter(Boolean)
  if (words.length === 0) {
    return 0
  }

  let lines = 1
  let used = 0
  for (const word of words) {
    const next = used === 0 ? word.length : used + 1 + word.length
    if (next > maxChars && used > 0) {
      lines += 1
      used = word.length
    } else {
      used = next
    }
  }

  return lines
}

export function estimatePersonHeight (person, style, width = CARD_WIDTH) {
  const textWidth = width - PAD_X - (style.showAvatars ? AVATAR + AVATAR_GAP : 0)
  const maxChars = Math.max(8, Math.floor(textWidth / CHAR_WIDTH))
  const nameLines = Math.max(1, wrapCount(person.fullName, maxChars))
  const titleLines = style.showJobTitles ? wrapCount(person.jobTitle, maxChars) : 0
  const teamLines = style.showTeam ? wrapCount(person.team, maxChars) : 0
  const textHeight = nameLines * NAME_LINE + titleLines * TITLE_LINE + teamLines * TEAM_LINE
  const body = Math.max(textHeight, style.showAvatars ? AVATAR : 0)

  return Math.ceil(body + PAD_Y)
}

export function measurePersonHeights (people, style, width = CARD_WIDTH, measure) {
  if (measure) {
    return measure(people, style, width)
  }

  const heights = new Map()
  for (const person of people) {
    heights.set(cardId(person), estimatePersonHeight(person, style, width))
  }
  return heights
}

export function cardHeights (people, style, compact, width = CARD_WIDTH, measure) {
  return equalizeRowHeights(measurePersonHeights(people, style, width, measure), cardRows(people, compact))
}
