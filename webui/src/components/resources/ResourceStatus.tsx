import { Flex, styled, Text } from '@traefiklabs/faency'
import { ReactNode } from 'react'

import { colorByStatus, iconByStatus, StatusType } from 'components/resources/Status'

export const StatusWrapper = styled(Flex, {
  height: '32px',
  width: '32px',
  padding: 0,
  borderRadius: '4px',
})

type Props = {
  status: StatusType
  label?: string
  withLabel?: boolean
}

type Value = { color: string; icon: ReactNode; label: string }

export const ResourceStatus = ({ status, withLabel = false }: Props) => {
  const valuesByStatus: { [key in StatusType]: Value } = {
    info: {
      color: colorByStatus.info,
      icon: iconByStatus.info,
      label: 'Info',
    },
    success: {
      color: colorByStatus.success,
      icon: iconByStatus.success,
      label: 'Success',
    },
    warning: {
      color: colorByStatus.warning,
      icon: iconByStatus.warning,
      label: 'Warning',
    },
    error: {
      color: colorByStatus.error,
      icon: iconByStatus.error,
      label: 'Error',
    },
    enabled: {
      color: colorByStatus.enabled,
      icon: iconByStatus.enabled,
      label: 'Success',
    },
    disabled: {
      color: colorByStatus.disabled,
      icon: iconByStatus.disabled,
      label: 'Error',
    },
  }

  const values = valuesByStatus[status]

  if (!values) {
    return null
  }

  return (
    <Flex css={{ alignItems: 'center' }} data-testid={status}>
      <StatusWrapper css={{ alignItems: 'center', justifyContent: 'center', backgroundColor: values.color }}>
        {values.icon}
      </StatusWrapper>
      {withLabel && values.label && (
        <Text css={{ ml: '$2', color: values.color, fontWeight: 600 }}>{values.label}</Text>
      )}
    </Flex>
  )
}
