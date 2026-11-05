import { Badge } from '@traefik-labs/faency'
import { useContext } from 'react'

import TooltipText from 'components/TooltipText'
import { VersionContext } from 'contexts/version'

const InstanceBadge = () => {
  const { dashboardName } = useContext(VersionContext)

  if (!dashboardName) return null

  return (
    <Badge size="small" data-testid="instance-badge" css={{ maxWidth: '100%', userSelect: 'none' }}>
      <TooltipText text={dashboardName} isTruncated css={{ color: 'inherit', fontSize: 'inherit', minWidth: 0 }} />
    </Badge>
  )
}

export default InstanceBadge
