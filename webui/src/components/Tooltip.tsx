import { Flex, Text } from '@traefiklabs/faency'
import { JSXElementConstructor, MouseEvent, ReactElement, ReactNode, useMemo } from 'react'
import { FiCopy } from 'react-icons/fi'

import { Button, Tooltip as FaencyTooltip } from './FaencyOverrides'

type TooltipProps = {
  action?: 'copy'
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  children: ReactNode & ReactElement<any, string | JSXElementConstructor<any>>
  label: string
}

export const Tooltip = ({ action, children, label }: TooltipProps) => {
  const actionComponent = useMemo(() => {
    if (action === 'copy') {
      return (
        <Button
          css={{ padding: '0 $2 !important' }}
          onClick={(e: MouseEvent): void => {
            e.stopPropagation()
            navigator.clipboard.writeText(label)
          }}
        >
          <FiCopy size={16} />
        </Button>
      )
    }

    return null
  }, [action, label])

  return (
    <FaencyTooltip
      content={
        <Flex align="center" gap={2} css={{ px: '$1' }}>
          <Text css={{ maxWidth: '240px !important', color: '$contrast', wordBreak: 'break-word' }}>{label}</Text>{' '}
          {actionComponent}
        </Flex>
      }
    >
      {children}
    </FaencyTooltip>
  )
}

export default Tooltip
