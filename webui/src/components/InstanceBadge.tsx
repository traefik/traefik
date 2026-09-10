import { Badge } from '@traefik-labs/faency'
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
    <Badge
      size="small"
      data-testid="instance-badge"
      aria-label={`Instance: ${dashboardName}`}
      css={{ maxWidth: '100%', userSelect: 'none' }}
    >
      {display}
    </Badge>
  )
}

export default InstanceBadge
