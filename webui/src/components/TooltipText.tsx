import { Text } from '@traefiklabs/faency'
import { useMemo } from 'react'

import Tooltip from 'components/Tooltip'

type TooltipTextProps = {
  isTruncated?: boolean
  text?: string
  variant?: 'short' | 'wide'
}

export default function TooltipText({ isTruncated = false, text, variant }: TooltipTextProps) {
  const maxWidth = useMemo(() => {
    switch (variant) {
      case 'short':
        return '60px'
      case 'wide':
        return '304px'
      default:
        return '208px'
    }
  }, [variant])

  const css = useMemo(
    () => (isTruncated ? { whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth } : undefined),
    [isTruncated, maxWidth],
  )

  if (typeof text === 'undefined') return <Text>-</Text>

  return (
    <Tooltip label={text} action="copy">
      <Text css={css}>{text}</Text>
    </Tooltip>
  )
}
