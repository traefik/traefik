import { Button, Flex, Text, Tooltip as FaencyTooltip } from '@traefiklabs/faency'
import { MouseEvent, ReactNode, useMemo } from 'react'
import { FiCopy } from 'react-icons/fi'

// FIXME content props type
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const CustomTooltip = FaencyTooltip as any

type TooltipProps = {
  action?: 'copy'
  children: ReactNode
  label: string
}

export default function Tooltip({ action, children, label }: TooltipProps) {
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
    <CustomTooltip
      content={
        <Flex align="center" gap={2} css={{ px: '$1' }}>
          <Text css={{ maxWidth: '240px !important', color: '$contrast', wordBreak: 'break-word' }}>{label}</Text>{' '}
          {actionComponent}
        </Flex>
      }
    >
      {children}
    </CustomTooltip>
  )
}
