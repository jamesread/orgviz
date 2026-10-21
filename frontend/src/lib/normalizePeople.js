// d3-org-chart needs exactly one root. People with no manager are attached
// to the first person, which is treated as the root.
export function normalizePeople (people) {
  const normalized = people.map((person) => ({ ...person }))

  for (const person of normalized) {
    if (person.parentId === 0) {
      person.parentId = 1
    }
    if (!person.avatarUrl) {
      person.avatarUrl = '/avatars/default.png'
    }
  }

  if (normalized.length > 0) {
    normalized[0].parentId = null
  }

  return normalized
}
