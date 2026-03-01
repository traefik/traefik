import { Button, Flex, Text, Tooltip as FaencyTooltip } from '@traefiklabs/faency'
import { MouseEvent, ReactNode, useMemo, useState } from 'react'
import { FiCheck, FiCopy } from 'react-icons/fi'

type TooltipProps = {
  action?: 'copy'
  children: ReactNode
  label: string
}

export default function Tooltip({ action, children, label }: TooltipProps) {
  const [showConfirmation, setShowConfirmation] = useState(false)

  const actionComponent = useMemo(() => {
    if (action === 'copy') {
      return (
        <Button
          css={{ padding: '0 $2 !important' }}
          onClick={async (e: MouseEvent) => {
            e.preventDefault()
            e.stopPropagation()
            await navigator.clipboard.writeText(label)
            setShowConfirmation(true)
            setTimeout(() => setShowConfirmation(false), 1500)
          }}
        >
          {showConfirmation ? <FiCheck size={16} /> : <FiCopy size={16} />}
        </Button>
      )
    }

    return null
  }, [action, label, showConfirmation])

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
