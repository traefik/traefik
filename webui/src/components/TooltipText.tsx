import { CSS, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'

import Tooltip from 'components/Tooltip'

type TooltipTextProps = {
  isTruncated?: boolean
  text?: string
  css?: CSS
}

export default function TooltipText({ isTruncated = false, text, css }: TooltipTextProps) {
  const appliedCss = useMemo(
    () =>
      isTruncated
        ? { whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: '100%', ...css }
        : css,
    [isTruncated, css],
  )

  if (typeof text === 'undefined') return <Text>-</Text>

  return (
    <Tooltip label={text} action="copy">
      <Text css={appliedCss}>{text}</Text>
    </Tooltip>
  )
}
