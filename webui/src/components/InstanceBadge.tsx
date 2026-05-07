import { Box, Text } from '@traefik-labs/faency'
import { useContext, useMemo } from 'react'

import { VersionContext } from 'contexts/version'

const MAX_LENGTH = 32

const truncate = (value: string): string => {
  const chars = Array.from(value)
  if (chars.length <= MAX_LENGTH) return value
  return chars.slice(0, MAX_LENGTH).join('') + '…'
}

const InstanceBadge = () => {
  const { dashboardName } = useContext(VersionContext)

  const display = useMemo(() => truncate(dashboardName), [dashboardName])

  if (!dashboardName) return null

  return (
    <Box
      data-testid="instance-badge"
      aria-label={`Instance: ${dashboardName}`}
      css={{
        backgroundColor: '$grayBlue4',
        borderRadius: '4px',
        px: '$2',
        py: '1px',
        ml: '$1',
        alignSelf: 'center',
        userSelect: 'none',
      }}
    >
      <Text
        css={{
          color: '$primary',
          fontWeight: '$semiBold',
          fontSize: '11px',
          textTransform: 'uppercase',
          letterSpacing: '0.5px',
          lineHeight: 1.4,
        }}
      >
        {display}
      </Text>
    </Box>
  )
}

export default InstanceBadge
