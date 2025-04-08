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
  NavigationItem,
  NavigationLink,
  styled,
  Text,
} from '@traefiklabs/faency'
import { ComponentProps, ReactNode, useMemo } from 'react'
import { Helmet } from 'react-helmet-async'
import { FiBookOpen, FiGithub, FiHelpCircle } from 'react-icons/fi'
import { matchPath, useNavigate } from 'react-router'
import { useLocation } from 'react-router-dom'
import useSWR from 'swr'

import Container from './Container'

import Logo from 'components/icons/Logo'
import { PluginsIcon } from 'components/icons/PluginsIcon'
import ThemeSwitcher from 'components/ThemeSwitcher'
import useTotals from 'hooks/use-overview-totals'
import routes, { NavRouteType } from 'routes'

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

const SubNavDrawer = styled(Flex, {
  boxShadow: '0 1px 3px rgba(0,0,0,0.2), 0 1px 1px rgba(0,0,0,0.14), 0 2px 1px -1px rgba(0,0,0,0.12)',
  mt: 64,
})

const SubNavContainer = styled(Container, {
  gap: '$2',
  py: '$2',
})

const TextHideOnTablet = styled(Text, {
  fontWeight: '$semiBold',

  '@media (min-width: ${breakpoints.tablet}) and (max-width: 1230px)': {
    display: 'none',
  },
})

const IconContainer = styled(Flex, {
  mr: 8,

  '@media (min-width: ${breakpoints.tablet}) and (max-width: ${breakpoints.laptop})': {
    mr: 0,
  },
})

interface NavBarItemProps extends NavRouteType {
  isActive: boolean
  isDisabled?: boolean
}

const NavBarItem = ({ path, label, icon, isActive, isDisabled = false }: NavBarItemProps) => {
  const navigate = useNavigate()

  return (
    <NavItemWithIcon
      isActive={isActive}
      onClick={(): false | void => !isDisabled && navigate(path)}
      icon={icon}
      label={label}
    />
  )
}

type NavItemWithIconProps = ComponentProps<typeof NavigationItem> &
  React.ButtonHTMLAttributes<HTMLButtonElement> & {
    label: string
    icon?: ReactNode
    isActive?: boolean
  }

const NavItemWrapper = ({
  externalLink,
  children,
  ...props
}: ComponentProps<typeof NavigationItem> & { externalLink?: string; children: ReactNode }) => {
  if (externalLink)
    return (
      <NavigationLink
        href={externalLink}
        target="_blank"
        css={{ mt: 0, '> div:first-child': { alignItems: 'center' } }}
        {...props}
      >
        {children}
      </NavigationLink>
    )

  return (
    <NavigationItem css={{ mt: 0, '> div:first-child': { alignItems: 'center' } }} {...props}>
      {children}
    </NavigationItem>
  )
}

const NavItemWithIcon = ({ icon, label, isActive, href, ...props }: NavItemWithIconProps & { href?: string }) => {
  return (
    <NavItemWrapper externalLink={href} active={isActive} {...props}>
      <IconContainer>
        <Text css={{ color: isActive ? '$primary' : '$navButtonText', opacity: isActive ? 1 : 0.74 }}>
          {icon ? icon : null}
        </Text>
      </IconContainer>
      <TextHideOnTablet css={{ color: isActive ? '$primary' : '$navButtonText', opacity: isActive ? 1 : 0.74 }}>
        {label}
      </TextHideOnTablet>
    </NavItemWrapper>
  )
}

const Header = () => {
  const location = useLocation()
  const navigate = useNavigate()

  const { data: version } = useSWR('/version')

  let currentSubRoute: NavRouteType | undefined

  const currentRoute: NavRouteType | undefined = routes.find((r) => {
    const pathMatcher = (route: NavRouteType) => matchPath(route.path, location.pathname)

    const mainMatch = pathMatcher(r)

    if (mainMatch) {
      return true
    } else if (r.subRoutes) {
      const srMatch = r.subRoutes.find((sr) => {
        const subMatch = pathMatcher(sr)

        if (subMatch) {
          return true
        } else if (sr.subPaths) {
          return sr.subPaths.some((path) => matchPath(path, location.pathname))
        }

        return false
      })

      if (srMatch) {
        currentSubRoute = srMatch
        return true
      }
    }

    return false
  })

  const showAdButton = useMemo(() => {
    if (!version) return false
    return !version?.disableDashboardAd
  }, [version])

  const useTotalsConfByPath: { [key: string]: { protocol: string } } = {
    '/http/': { protocol: 'http' },
    '/tcp/': { protocol: 'tcp' },
    '/udp/': { protocol: 'udp' },
  }
  const useTotalsConf = currentRoute && useTotalsConfByPath[currentRoute.path]
  const { routers, services, middlewares } = useTotals(useTotalsConf ? useTotalsConf : { enabled: false })

  const totalValueByPath: { [key: string]: number } = {
    '/http/routers': routers,
    '/http/services': services,
    '/http/middlewares': middlewares,
    '/tcp/routers': routers,
    '/tcp/services': services,
    '/tcp/middlewares': middlewares,
    '/udp/routers': routers,
    '/udp/services': services,
  }

  return (
    <>
      {showAdButton && (
        <Helmet>
          <script src="https://traefik.github.io/traefiklabs-hub-button-app/main-v1.js"></script>
        </Helmet>
      )}
      <NavigationDrawer css={{ position: 'fixed', zIndex: 999, top: 0 }}>
        <NavigationContainer css={{ overflowX: 'auto' }}>
          <Flex align="center" justify="space-between" css={{ flexGrow: 1 }}>
            <Flex align="center" gap={2}>
              <Flex
                as="a"
                align="center"
                gap={2}
                css={{ color: '$primary', mt: '$1', mr: '$2', textDecoration: 'none', minWidth: 'fit-content' }}
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

              {routes.map((route) => (
                <NavBarItem {...route} key={route.path} isActive={currentRoute?.path === route.path} />
              ))}
              <NavItemWithIcon
                href="https://plugins.traefik.io/"
                target="_blank"
                icon={<PluginsIcon />}
                label="Plugins"
              />
            </Flex>
            <Flex gap={2} align="center" css={{ ml: '$4' }}>
              {showAdButton && (
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
          </Flex>
        </NavigationContainer>
      </NavigationDrawer>
      {currentRoute?.subRoutes && (
        <SubNavDrawer data-testid="subnavbar">
          <SubNavContainer>
            {currentRoute.subRoutes.map((route) => (
              <div key={route.path}>
                <NavigationItem
                  onClick={(): void => navigate(route.path)}
                  active={currentSubRoute?.path === route.path}
                >
                  <Flex align="center">
                    {route.label}
                    {!!totalValueByPath[route.path] && (
                      <Badge variant={currentSubRoute?.path === route.path ? 'green' : undefined} css={{ ml: '$2' }}>
                        {totalValueByPath[route.path]}
                      </Badge>
                    )}
                  </Flex>
                </NavigationItem>
              </div>
            ))}
          </SubNavContainer>
        </SubNavDrawer>
      )}
    </>
  )
}

export default Header
