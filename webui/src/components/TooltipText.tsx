import { Text } from '@traefiklabs/faency'
import { useMemo } from 'react'

import Tooltip from 'components/Tooltip'

type TooltipTextProps = {
  isTruncated?: boolean
  text?: string
}

export default function TooltipText({ isTruncated = false, text }: TooltipTextProps) {
  const css = useMemo(
    () =>
      isTruncated
        ? { whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: '100%' }
        : undefined,
    [isTruncated],
  )

  if (typeof text === 'undefined') return <Text>-</Text>

  return (
    <Tooltip label={text} action="copy">
      <Text css={css}>{text}</Text>
    </Tooltip>
  )
}
