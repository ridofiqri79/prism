const allowedTags = new Set(['B', 'BR', 'DIV', 'EM', 'I', 'LI', 'OL', 'P', 'SPAN', 'STRONG', 'U', 'UL'])
const droppedTags = new Set(['IFRAME', 'OBJECT', 'SCRIPT', 'STYLE'])

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
    .replace(/\r?\n/g, '<br>')
}

function textNodeToFragment(text: string) {
  const fragment = document.createDocumentFragment()
  const lines = text.split(/\r?\n/)

  lines.forEach((line, index) => {
    if (index > 0) {
      fragment.appendChild(document.createElement('br'))
    }

    if (line) {
      fragment.appendChild(document.createTextNode(line))
    }
  })

  return fragment
}

export function sanitizeRichText(value?: string | null) {
  if (!value) return ''

  if (typeof document === 'undefined') {
    return escapeHtml(value)
  }

  const template = document.createElement('template')
  template.innerHTML = value

  function sanitizeNode(node: Node): Node | null {
    if (node.nodeType === Node.TEXT_NODE) {
      const text = node.textContent ?? ''
      if (text.includes('\n')) {
        return textNodeToFragment(text)
      }

      return document.createTextNode(text)
    }

    if (node.nodeType !== Node.ELEMENT_NODE) {
      return null
    }

    const element = node as HTMLElement
    const tagName = element.tagName.toUpperCase()

    if (droppedTags.has(tagName)) {
      return null
    }

    if (!allowedTags.has(tagName)) {
      const fragment = document.createDocumentFragment()
      element.childNodes.forEach((child) => {
        const sanitized = sanitizeNode(child)
        if (sanitized) fragment.appendChild(sanitized)
      })
      return fragment
    }

    const normalizedTag = tagName === 'DIV' || tagName === 'SPAN' ? 'p' : tagName.toLowerCase()
    const clean = document.createElement(normalizedTag)
    element.childNodes.forEach((child) => {
      const sanitized = sanitizeNode(child)
      if (sanitized) clean.appendChild(sanitized)
    })
    return clean
  }

  const output = document.createElement('div')
  template.content.childNodes.forEach((node) => {
    const sanitized = sanitizeNode(node)
    if (sanitized) output.appendChild(sanitized)
  })

  return output.innerHTML.trim()
}

export function isRichTextEmpty(value?: string | null) {
  const sanitized = sanitizeRichText(value)
  if (!sanitized) return true

  if (typeof document === 'undefined') {
    return sanitized.replace(/<[^>]*>/g, '').trim() === ''
  }

  const wrapper = document.createElement('div')
  wrapper.innerHTML = sanitized
  return (wrapper.textContent ?? '').trim() === ''
}
