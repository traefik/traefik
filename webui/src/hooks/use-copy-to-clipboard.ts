import { useCallback } from 'react'

// Falls back to execCommand for non-secure contexts (e.g. plain HTTP), where
// navigator.clipboard is unavailable.
const copyWithFallback = (text: string): boolean => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'

  try {
    document.body.appendChild(textarea)
    textarea.select()
    return document.execCommand('copy')
  } catch {
    return false
  } finally {
    textarea.remove()
  }
}

const useCopyToClipboard = (): ((text: string) => Promise<boolean>) => {
  return useCallback(async (text: string): Promise<boolean> => {
    if (navigator.clipboard) {
      try {
        await navigator.clipboard.writeText(text)
        return true
      } catch {
        return copyWithFallback(text)
      }
    }

    return copyWithFallback(text)
  }, [])
}

export default useCopyToClipboard
