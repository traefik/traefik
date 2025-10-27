import { Badge, Box, Flex, Text } from '@traefiklabs/faency'
import { useContext, useState } from 'react'
import { BsChevronRight } from 'react-icons/bs'

import { HubDemoContext } from './demoNavContext'
import { HubIcon } from './icons'

import Tooltip from 'components/Tooltip'
import { BasicNavigationItem, LAPTOP_BP } from 'layout/Navigation'

const ApimDemoNavMenu = ({
  isResponsive,
  isSmallScreen,
  isExpanded,
}: {
  isResponsive: boolean
  isSmallScreen: boolean
  isExpanded: boolean
}) => {
  const [isCollapsed, setIsCollapsed] = useState(false)
  const { navigationItems: hubDemoNavItems } = useContext(HubDemoContext)

  if (!hubDemoNavItems) {
    return null
  }

  return (
    <Flex direction="column" css={{ borderTop: '1px solid $colors$tableRowBorder', borderRadius: 0, pt: '$3' }}>
      <Flex
        align="center"
        css={{ color: '$grayBlue9', my: '$2', cursor: 'pointer' }}
        onClick={() => setIsCollapsed(!isCollapsed)}
      >
        <BsChevronRight
          size={12}
          style={{
            transform: isCollapsed ? 'rotate(90deg)' : 'unset',
            transition: 'transform 0.3s ease-in-out',
          }}
        />
        {isSmallScreen ? (
          <Tooltip label="Hub demo">
            <Box css={{ ml: 4, color: '$navButtonText' }}>
              <HubIcon width={20} />
            </Box>
          </Tooltip>
        ) : (
          <>
            <Text
              css={{
                fontWeight: 600,
                textTransform: 'uppercase',
                letterSpacing: 0.2,
                color: '$grayBlue9',
                ml: 8,
                [`@media (max-width:${LAPTOP_BP}px)`]: isResponsive ? { display: 'none' } : undefined,
              }}
            >
              API management
            </Text>
            <Badge variant="green" css={{ ml: '$2' }}>
              Demo
            </Badge>
          </>
        )}
      </Flex>

      <Box
        css={{ mt: '$1', transition: 'max-height 0.3s ease-out', maxHeight: isCollapsed ? 500 : 0, overflow: 'hidden' }}
      >
        {hubDemoNavItems.map((route, idx) => (
          <BasicNavigationItem
            key={`apim-${idx}`}
            route={route}
            isSmallScreen={isSmallScreen}
            isExpanded={isExpanded}
          />
        ))}
      </Box>
    </Flex>
  )
}

export default ApimDemoNavMenu
