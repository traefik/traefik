import { ReactNode, useEffect, useMemo, useState } from 'react'
import { RouteObject } from 'react-router-dom'

import HubDashboard from 'pages/hub-demo/HubDashboard'
import { ApiIcon, DashboardIcon, GatewayIcon, PortalIcon } from 'pages/hub-demo/icons'
import verifySignature from 'pages/hub-demo/workers/scriptVerification'

const ROUTES_MANIFEST_URL = 'https://traefik.github.io/hub-ui-demo-app/config/routes.json'

const HUB_DEMO_NAV_ICONS: Record<string, ReactNode> = {
  dashboard: <DashboardIcon color="currentColor" width={22} height={22} />,
  gateway: <GatewayIcon color="currentColor" width={22} height={22} />,
  api: <ApiIcon color="currentColor" width={22} height={22} />,
  portal: <PortalIcon color="currentColor" width={22} height={22} />,
}

const useHubDemoRoutesManifest = (): HubDemo.Manifest | null => {
  const [manifest, setManifest] = useState<HubDemo.Manifest | null>(null)

  useEffect(() => {
    const fetchManifest = async () => {
      try {
        const { verified, scriptContent } = await verifySignature(ROUTES_MANIFEST_URL, `${ROUTES_MANIFEST_URL}.sig`)

        if (!verified || !scriptContent) {
          setManifest(null)
          return
        }

        const textDecoder = new TextDecoder()
        const jsonString = textDecoder.decode(scriptContent)
        const data: HubDemo.Manifest = JSON.parse(jsonString)
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

export const useHubDemo = (basePath: string) => {
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

  const navigationItems = useMemo(() => {
    if (!manifest) {
      return null
    }

    return manifest.routes.map((route) => ({
      path: `${basePath}${route.path}`,
      label: route.label,
      icon: HUB_DEMO_NAV_ICONS[route.icon],
      activeMatches: route.activeMatches?.map((r) => `${basePath}${r}`),
    }))
  }, [basePath, manifest])

  return { routes, navigationItems }
}
