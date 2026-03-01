import { Button } from '@traefiklabs/faency'
import { useCallback, useEffect, useState } from 'react'

export const ScrollTopButton = () => {
  const [showOnScroll, setShowOnScroll] = useState<boolean>(false)

  const handleScroll = useCallback(() => {
    const position = window?.scrollY || 0
    setShowOnScroll(position >= 160)
  }, [setShowOnScroll])

  useEffect(() => {
    window.addEventListener('scroll', handleScroll, { passive: true })

    return () => {
      window.removeEventListener('scroll', handleScroll)
    }
  }, [handleScroll])

  if (!showOnScroll) {
    return null
  }

  return (
    <Button variant="primary" onClick={(): void => window.scrollTo({ top: 0, behavior: 'smooth' })}>
      Scroll to top
    </Button>
  )
}
