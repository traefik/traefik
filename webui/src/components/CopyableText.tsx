import { Text } from '@traefiklabs/faency'
import { useContext } from 'react'

import CopyButton from 'components/CopyButton'
import { ToastContext } from 'contexts/toasts'

type CopyableTextProps = {
  notifyText?: string
  text: string
}

export default function CopyableText({ notifyText, text }: CopyableTextProps) {
  const { addToast } = useContext(ToastContext)

  return (
    <Text
      css={{
        flex: '1 1 auto',
        minWidth: 0,
        overflow: 'hidden',
        overflowWrap: 'anywhere',
        verticalAlign: 'middle',
        fontSize: 'inherit',
      }}
    >
      {text}
      <CopyButton
        text={text}
        onClick={() => {
          if (notifyText) addToast({ message: notifyText, severity: 'success' })
        }}
        css={{ display: 'inline-block', height: 20, verticalAlign: 'middle', ml: '$1' }}
        iconOnly
      />
    </Text>
  )
}
