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
  styled,
  Text,
} from '@traefiklabs/faency'
import { useMemo } from 'react'
import { FiBookOpen, FiGithub, FiHelpCircle } from 'react-icons/fi'
import { matchPath, useHref } from 'react-router'
import { useLocation } from 'react-router-dom'

import Container from './Container'

import Logo from 'components/icons/Logo'
import { PluginsIcon } from 'components/icons/PluginsIcon'
import ThemeSwitcher from 'components/ThemeSwitcher'
import useTotals from 'hooks/use-overview-totals'
import useVersion from 'hooks/use-version'
import { Route, ROUTES } from 'routes'

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

const NavigationContainer = Container

const BasicNavigationItem = ({ route, count }: { route: Route; count?: number }) => {
  const { pathname } = useLocation()
  const href = useHref(route.path)

  const isActiveRoute = useMemo(() => {
    const mainPath = matchPath(route.path, pathname)

    if (mainPath) return true

    if (route.activeMatches) {
      return route.activeMatches.some((path) => matchPath(path, pathname))
    }
  }, [pathname, route.activeMatches, route.path])

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

export const SideNav = () => {
  const { version } = useVersion()

  const { http, tcp, udp } = useTotals()

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
      css={{
        width: 260,
        height: '100vh',
        flexDirection: 'column',
      }}
    >
      <NavigationContainer
        css={{
          overflow: 'auto',
          py: '$3',
          flexDirection: 'column',
        }}
        data-testid="nav-container"
      >
        <Flex
          as="a"
          align="center"
          gap={2}
          css={{ color: '$primary', mt: '$3', mb: '$6', textDecoration: 'none', height: 'fit-content' }}
          href="https://github.com/traefik/traefik/"
          target="_blank"
        >
          <Logo height={32} />
          {!!version && (
            <Text variant="subtle" size="4" css={{ fontWeight: '$semiBold' }}>
              {version.Version}
            </Text>
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
          }}
          href="https://plugins.traefik.io/"
          target="_blank"
        >
          Plugins
        </NavigationLink>
      </NavigationContainer>
    </NavigationDrawer>
  )
}

export const TopNav = () => {
  const { showHubButton } = useVersion()

  return (
    <Flex as="nav" role="navigation" justify="end" align="center" css={{ gap: '$2', mb: '$6' }}>
      {showHubButton && (
        <Box css={{ fontFamily: '$rubik', fontWeight: '500 !important' }}>
          <hub-button-app />
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
                  href={import.meta.env.VITE_APP_DOCS_URL}
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
                  href={import.meta.env.VITE_APP_REPO_URL}
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
