import { Box, Flex, styled, Text } from '@traefiklabs/faency'
import { ReactNode } from 'react'

import { colorByStatus, iconByStatus } from 'components/resources/Status'

export const StatusWrapper = styled(Flex, {
  height: '24px',
  width: '24px',
  padding: 0,
  borderRadius: '4px',
})

type Props = {
  status: Resource.Status
  label?: string
  withLabel?: boolean
  size?: number
}

type Value = { color: string; icon: ReactNode; label: string }

export const ResourceStatus = ({ status, withLabel = false, size = 20 }: Props) => {
  const valuesByStatus: { [key in Resource.Status]: Value } = {
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
    loading: {
      color: colorByStatus.loading,
      icon: iconByStatus.loading,
      label: 'Loading...',
    },
  }

  const values = valuesByStatus[status]

  if (!values) {
    return null
  }

  return (
    <Flex align="center" css={{ width: size, height: size }} data-testid={status}>
      <Box css={{ color: values.color, width: size, height: size }}>{values.icon}</Box>
      {withLabel && values.label && (
        <Text css={{ ml: '$2', color: values.color, fontWeight: 600, fontSize: 'inherit !important' }}>
          {values.label}
        </Text>
      )}
    </Flex>
  )
}
