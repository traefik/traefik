import { Text } from '@traefiklabs/faency'

import Tooltip from 'components/Tooltip'

export default function TruncatedText({ maxWidth = '64px', text }: { maxWidth?: string; text: string }) {
  return (
    <Tooltip label={text} action="copy">
      <Text css={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth }}>{text}</Text>
    </Tooltip>
  )
}
