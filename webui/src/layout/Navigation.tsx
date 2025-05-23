import {
  Badge,
  Box,
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuTrigger,
  elevationVariants,
  Flex,
  Link,
  NavigationLink,
  SidePanel,
  styled,
  Text,
  Tooltip,
} from '@traefiklabs/faency'
import { useEffect, useMemo, useState } from 'react'
import { BsChevronDoubleRight, BsChevronDoubleLeft } from 'react-icons/bs'
import { FiBookOpen, FiGithub, FiHelpCircle } from 'react-icons/fi'
import { matchPath, useHref } from 'react-router'
import { useLocation } from 'react-router-dom'
import { useWindowSize } from 'usehooks-ts'

import Container from './Container'

import IconButton from 'components/buttons/IconButton'
import Logo from 'components/icons/Logo'
import { PluginsIcon } from 'components/icons/PluginsIcon'
import ThemeSwitcher from 'components/ThemeSwitcher'
import TooltipText from 'components/TooltipText'
import useTotals from 'hooks/use-overview-totals'
import useVersion from 'hooks/use-version'
import { Route, ROUTES } from 'routes'

export const LAPTOP_BP = 1025

const NavigationDrawer = styled(Flex, {
  width: '100%',
  maxWidth: '100%',
  height: 64,
  p: 0,
  variants: {
    elevation: elevationVariants,
  },
  defaultVariants: {
    elevation: 1,
  },
})

const BasicNavigationItem = ({
  route,
  count,
  isSmallScreen,
  isExpanded,
}: {
  route: Route
  count?: number
  isSmallScreen: boolean
  isExpanded: boolean
}) => {
  const { pathname } = useLocation()
  const href = useHref(route.path)

  const isActiveRoute = useMemo(() => {
    const mainPath = matchPath(route.path, pathname)

    if (mainPath) return true

    if (route.activeMatches) {
      return route.activeMatches.some((path) => matchPath(path, pathname))
    }
  }, [pathname, route.activeMatches, route.path])

  if (isSmallScreen && !isExpanded) {
    return (
      <Tooltip content={<Text css={{ color: '$tooltipText' }}>{route.label}</Text>} side="right">
        <Box>
          <NavigationLink active={isActiveRoute} startAdornment={route?.icon} href={href} />
        </Box>
      </Tooltip>
    )
  }

  return (
    <NavigationLink active={isActiveRoute} startAdornment={route?.icon} css={{ whiteSpace: 'nowrap' }} href={href}>
      {route.label}
      {!!count && (
        <Badge variant={isActiveRoute ? 'green' : undefined} css={{ ml: '$2' }}>
          {count}
        </Badge>
      )}
    </NavigationLink>
  )
}

export const SideBarPanel = ({
  isOpen,
  onOpenChange,
}: {
  isOpen: boolean
  onOpenChange: (isOpen: boolean) => void
}) => {
  const windowSize = useWindowSize()

  return (
    <SidePanel
      open={isOpen && windowSize.width < LAPTOP_BP}
      onOpenChange={onOpenChange}
      side="left"
      css={{ width: 250, p: 0 }}
      description="Expanded side navigation"
      noCloseIcon
    >
      <SideNav isExpanded={isOpen} onSidePanelToggle={() => onOpenChange(false)} />
    </SidePanel>
  )
}

export const SideNav = ({
  isExpanded,
  onSidePanelToggle,
  isResponsive = false,
}: {
  isExpanded: boolean
  onSidePanelToggle: () => void
  isResponsive?: boolean
}) => {
  const windowSize = useWindowSize()
  const { version } = useVersion()

  const { http, tcp, udp } = useTotals()

  const [isSmallScreen, setIsSmallScreen] = useState(false)

  useEffect(() => {
    setIsSmallScreen(isResponsive && windowSize.width < LAPTOP_BP)
  }, [isExpanded, isResponsive, windowSize.width])

  const totalValueByPath = useMemo<{ [key: string]: number }>(
    () => ({
      '/http/routers': http?.routers,
      '/http/services': http?.services,
      '/http/middlewares': http?.middlewares as number,
      '/tcp/routers': tcp?.routers,
      '/tcp/services': tcp?.services,
      '/tcp/middlewares': tcp?.middlewares as number,
      '/udp/routers': udp?.routers,
      '/udp/services': udp?.services,
    }),
    [http, tcp, udp],
  )

  return (
    <NavigationDrawer
      data-collapsed={isExpanded && isResponsive && isSmallScreen}
      css={{
        width: 250,
        height: '100vh',
        flexDirection: 'column',
        position: 'fixed',
        [`@media (max-width:${LAPTOP_BP}px)`]: isResponsive
          ? {
              width: 60,
              px: '$2',
            }
          : undefined,
        transition: '150ms cubic-bezier(0.22, 1, 0.36, 1)',
        '&[data-collapsed="true"]': {
          marginLeft: -32,
        },
      }}
    >
      <IconButton
        ghost
        icon={isExpanded ? <BsChevronDoubleLeft size={16} /> : <BsChevronDoubleRight size={16} />}
        onClick={onSidePanelToggle}
        css={{
          display: 'none',
          position: 'absolute',
          top: 3,
          right: 0,
          color: '$hiContrast',
          [`@media (max-width:${LAPTOP_BP}px)`]: { display: 'inherit' },
          p: '$1',
          '&:before, &:after': { borderRadius: '10px' },
          height: 16,
        }}
      />
      <Container
        css={{
          overflow: 'auto',
          py: '$3',
          flexDirection: 'column',
          mx: '$3',
          [`@media (max-width:${LAPTOP_BP}px)`]: isResponsive ? { m: 0 } : undefined,
        }}
        data-testid="nav-container"
      >
        <Flex
          as="a"
          gap={2}
          css={{
            color: '$primary',
            mt: '$3',
            mb: '$6',
            textDecoration: 'none',
            height: 'fit-content',
            pl: '$3',
            [`@media (max-width:${LAPTOP_BP}px)`]: isResponsive
              ? { mt: '$4', px: 0, justifyContent: 'center' }
              : undefined,
          }}
          href="https://github.com/traefik/traefik/"
          target="_blank"
          data-testid="proxy-main-nav"
        >
          <Logo height={isSmallScreen ? 36 : 56} isSmallScreen={isSmallScreen} />
          {!!version && !isSmallScreen && (
            <TooltipText text={version.Version} css={{ maxWidth: 50, fontWeight: '$semiBold' }} isTruncated />
          )}
        </Flex>
        {ROUTES.map((section, index) => (
          <Flex
            key={`nav-section-${index}`}
            direction="column"
            gap="1"
            css={{
              '&:not(:last-child)': {
                borderTop: section.sectionLabel ? '1px solid $colors$tableRowBorder' : undefined,
                mb: '$3',
                pt: section.sectionLabel ? '$3' : undefined,
              },
            }}
          >
            {section.sectionLabel && (
              <Text
                css={{
                  fontWeight: 600,
                  color: '$grayBlue9',
                  mb: '$2',
                  textTransform: 'uppercase',
                  letterSpacing: 0.2,
                  ml: 15,
                  [`@media (max-width:${LAPTOP_BP}px)`]: isResponsive ? { display: 'none' } : undefined,
                }}
              >
                {section.sectionLabel}
              </Text>
            )}
            {section.items.map((item, idx) => (
              <BasicNavigationItem
                key={`nav-section-${index}-${idx}`}
                route={item}
                count={totalValueByPath[item.path]}
                isSmallScreen={isSmallScreen}
                isExpanded={isExpanded}
              />
            ))}
          </Flex>
        ))}
        <NavigationLink
          startAdornment={<PluginsIcon />}
          css={{
            whiteSpace: 'nowrap',
            borderTop: '1px solid $colors$tableRowBorder',
            mb: '$3',
            pt: '$3',
            mt: 0,
            borderRadius: 0,
          }}
          href="https://plugins.traefik.io/"
          target="_blank"
        >
          {!isSmallScreen || isExpanded ? 'Plugins' : ''}
        </NavigationLink>
      </Container>
    </NavigationDrawer>
  )
}

export const TopNav = () => {
  const { showHubButton, version } = useVersion()

  const parsedVersion = useMemo(() => {
    if (!version?.Version) {
      return 'master'
    }
    if (version.Version === 'dev') {
      return 'master'
    }
    const matches = version.Version.match(/^(v?\d+\.\d+)/)
    return matches ? 'v' + matches[1] : 'master'
  }, [version])

  return (
    <Flex as="nav" role="navigation" justify="end" align="center" css={{ gap: '$2', mb: '$6' }}>
      {showHubButton && (
        <Box css={{ fontFamily: '$rubik', fontWeight: '500 !important' }}>
          <hub-button-app style={{ backgroundColor: '#2AA2C1' }} />
        </Box>
      )}
      <ThemeSwitcher />

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button ghost variant="secondary" css={{ px: '$2', boxShadow: 'none' }} data-testid="help-menu">
            <FiHelpCircle size={20} />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuPortal>
          <DropdownMenuContent align="end" css={{ zIndex: 9999 }}>
            <DropdownMenuGroup>
              <DropdownMenuItem css={{ height: '$6', cursor: 'pointer' }}>
                <Link
                  href={`https://doc.traefik.io/traefik/${parsedVersion}`}
                  target="_blank"
                  css={{ textDecoration: 'none', '&:hover': { textDecoration: 'none' } }}
                >
                  <Flex align="center" gap={2}>
                    <FiBookOpen size={20} />
                    <Text>Documentation</Text>
                  </Flex>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem css={{ height: '$6', cursor: 'pointer' }}>
                <Link
                  href="https://github.com/traefik/traefik/"
                  target="_blank"
                  css={{ textDecoration: 'none', '&:hover': { textDecoration: 'none' } }}
                >
                  <Flex align="center" gap={2}>
                    <FiGithub size={20} />
                    <Text>Github Repository</Text>
                  </Flex>
                </Link>
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenuPortal>
      </DropdownMenu>
    </Flex>
  )
}
