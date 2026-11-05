import { Box, Text } from '@traefik-labs/faency'
import { useContext } from 'react'

import Tooltip from 'components/Tooltip'
import { VersionContext } from 'contexts/version'

const InstanceBadge = () => {
  const { dashboardName } = useContext(VersionContext)

  if (!dashboardName) return null

  return (
    <Tooltip label={dashboardName} action="copy">
      <Box
        data-testid="instance-badge"
        css={{ width: 'fit-content', maxWidth: '100%', px: '$2', py: '$1', mb: '$3', borderRadius: '$2', bc: '$contentBg' }}
      >
        <Text
          css={{
            display: 'block',
            fontSize: '$2',
            lineHeight: 1.4,
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {dashboardName}
        </Text>
      </Box>
    </Tooltip>
  )
}

export default InstanceBadge
