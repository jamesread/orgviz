import { cardId } from './cardHeight.js'

export function nextSelectedId (currentId, clickedId) {
  const current = currentId == null ? '' : String(currentId)
  const clicked = clickedId == null ? '' : String(clickedId)
  if (!clicked || current === clicked) {
    return ''
  }

  return clicked
}

export function applyNodeSelection (root, selectedId) {
  if (!root) {
    return
  }

  const selected = selectedId == null ? '' : String(selectedId)

  for (const el of root.querySelectorAll('.person')) {
    const isSelected = selected !== '' && el.dataset.personId === selected
    el.classList.toggle('selected', isSelected)
    if (isSelected) {
      el.setAttribute('aria-current', 'true')
    } else {
      el.removeAttribute('aria-current')
    }
    el.closest('.node')?.classList.toggle('selected', isSelected)
  }
}

function clickedPersonId (target) {
  return target.closest('.person')?.dataset.personId || ''
}

function isChartChrome (target) {
  return Boolean(target.closest('.node-button-g, .paging-button-wrapper'))
}

export function bindChartSelection (root, state) {
  function select (id) {
    state.setSelectedId(id)
    applyNodeSelection(root, id)
  }

  function onClick (event) {
    if (isChartChrome(event.target)) {
      return
    }

    const id = clickedPersonId(event.target)
    if (id) {
      select(nextSelectedId(state.getSelectedId(), id))
      return
    }

    select('')
  }

  function onKey (event) {
    if (event.key === 'Escape') {
      select('')
    }
  }

  applyNodeSelection(root, state.getSelectedId())
  root.addEventListener('click', onClick)
  document.addEventListener('keydown', onKey)

  return () => {
    root.removeEventListener('click', onClick)
    document.removeEventListener('keydown', onKey)
  }
}

export function keepSelectedId (selectedId, people) {
  const id = selectedId == null ? '' : String(selectedId)
  if (!id) {
    return ''
  }

  const found = people.some((person) => cardId(person) === id)
  return found ? id : ''
}
