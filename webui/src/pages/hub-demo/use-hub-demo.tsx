import { ReactNode, useEffect, useMemo, useState } from 'react'
import { RouteObject } from 'react-router-dom'

import HubDashboard from 'pages/hub-demo/HubDashboard'
import { ApiIcon, DashboardIcon, GatewayIcon, PortalIcon } from 'pages/hub-demo/icons'
import verifySignature from 'pages/hub-demo/workers/scriptVerification'

const ROUTES_MANIFEST_SOURCE = 'https://traefik.github.io/hub-ui-demo-app/config/routes.json'

export const useHubDemoRoutesManifest = (): HubDemo.Manifest | null => {
  const [manifest, setManifest] = useState<HubDemo.Manifest | null>(null)

  useEffect(() => {
    const fetchManifest = async () => {
      try {
        const isSignatureValid = await verifySignature(ROUTES_MANIFEST_SOURCE, `${ROUTES_MANIFEST_SOURCE}.sig`)

        if (!isSignatureValid) {
          console.error('Manifest signature verification failed - security violation detected')
          setManifest(null)
          return
        }

        const response = await fetch(ROUTES_MANIFEST_SOURCE)

        if (!response.ok) {
          throw new Error(`Failed to fetch hub demo manifest: ${response.statusText}`)
        }

        const data: HubDemo.Manifest = await response.json()
        setManifest(data)
      } catch (error) {
        console.error('Failed to load hub demo manifest:', error)
        setManifest(null)
      }
    }

    fetchManifest()
  }, [])

  return manifest
}

export const useHubDemoRoutes = (basePath: string): RouteObject[] | null => {
  const manifest = useHubDemoRoutesManifest()

  const routes = useMemo(() => {
    if (!manifest) {
      return null
    }

    const routeObjects: RouteObject[] = []

    manifest.routes.forEach((route: HubDemo.Route) => {
      routeObjects.push({
        path: `${basePath}${route.path}`,
        element: <HubDashboard path={route.contentPath} />,
      })

      if (route.dynamicSegments) {
        route.dynamicSegments.forEach((segment) => {
          routeObjects.push({
            path: `${basePath}${route.path}/${segment}`,
            element: <HubDashboard path={`${route.contentPath}${segment}`} />,
          })
        })
      }
    })

    return routeObjects
  }, [basePath, manifest])

  return routes
}

const HUB_DEMO_NAV_ICONS: Record<string, ReactNode> = {
  dashboard: <DashboardIcon color="currentColor" width={22} height={22} />,
  gateway: <GatewayIcon color="currentColor" width={22} height={22} />,
  api: <ApiIcon color="currentColor" width={22} height={22} />,
  portal: <PortalIcon color="currentColor" width={22} height={22} />,
}

export const useHubDemoNavigation = (basePath: string): HubDemo.NavItem[] | null => {
  const manifest = useHubDemoRoutesManifest()

  const navItems = useMemo(() => {
    if (!manifest) {
      return null
    }

    return manifest.routes.map((route) => ({
      path: `${basePath}${route.path}`,
      label: route.label,
      icon: HUB_DEMO_NAV_ICONS[route.icon],
      activeMatches: route.activeMatches,
    }))
  }, [basePath, manifest])

  return navItems
}
