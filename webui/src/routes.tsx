import { ReactNode } from 'react'
import { LiaProjectDiagramSolid, LiaServerSolid, LiaCogsSolid, LiaHomeSolid } from 'react-icons/lia'

import { ApiIcon, DashboardIcon, GatewayIcon, PortalIcon } from 'pages/hub-demo/icons'

export type Route = {
  path: string
  label: string
  icon?: string | ReactNode
  activeMatches?: string[]
}

type RouteSections = {
  section: string
  items: Route[]
  sectionLabel?: string
}

export const ROUTES: RouteSections[] = [
  {
    section: 'dashboard',
    items: [
      {
        path: '/',
        label: 'Dashboard',
        icon: <LiaHomeSolid color="currentColor" size={20} />,
      },
    ],
  },
  {
    section: 'http',
    sectionLabel: 'HTTP',
    items: [
      {
        path: '/http/routers',
        activeMatches: ['/http/routers/:name'],
        label: 'HTTP Routers',
        icon: <LiaProjectDiagramSolid color="currentColor" size={20} />,
      },
      {
        path: '/http/services',
        activeMatches: ['/http/services/:name'],
        label: 'HTTP Services',
        icon: <LiaServerSolid color="currentColor" size={20} />,
      },
      {
        path: '/http/middlewares',
        activeMatches: ['/http/middlewares/:name'],
        label: 'HTTP Middlewares',
        icon: <LiaCogsSolid color="currentColor" size={20} />,
      },
    ],
  },
  {
    section: 'tcp',
    sectionLabel: 'TCP',
    items: [
      {
        path: '/tcp/routers',
        activeMatches: ['/tcp/routers/:name'],
        label: 'TCP Routers',
        icon: <LiaProjectDiagramSolid color="currentColor" size={20} />,
      },
      {
        path: '/tcp/services',
        activeMatches: ['/tcp/services/:name'],
        label: 'TCP Services',
        icon: <LiaServerSolid color="currentColor" size={20} />,
      },
      {
        path: '/tcp/middlewares',
        activeMatches: ['/tcp/middlewares/:name'],
        label: 'TCP Middlewares',
        icon: <LiaCogsSolid color="currentColor" size={20} />,
      },
    ],
  },
  {
    section: 'udp',
    sectionLabel: 'UDP',
    items: [
      {
        path: '/udp/routers',
        activeMatches: ['/udp/routers/:name'],
        label: 'UDP Routers',
        icon: <LiaProjectDiagramSolid color="currentColor" size={20} />,
      },
      {
        path: '/udp/services',
        activeMatches: ['/udp/services/:name'],
        label: 'UDP Services',
        icon: <LiaServerSolid color="currentColor" size={20} />,
      },
    ],
  },
]

export const APIM_DEMO_ROUTES = [
  {
    path: '/hub-dashboard',
    label: 'Dashboard',
    icon: <DashboardIcon color="currentColor" width={22} height={22} />,
  },
  {
    path: '/hub-dashboard/gateways',
    label: 'Gateways',
    icon: <GatewayIcon color="currentColor" width={22} height={22} />,
    activeMatches: ['/hub-dashboard/gateways/:id'],
  },
  {
    path: '/hub-dashboard/apis',
    label: 'APIs',
    icon: <ApiIcon color="currentColor" width={22} height={22} />,
    activeMatches: ['/hub-dashboard/apis/:id'],
  },
  {
    path: '/hub-dashboard/portals',
    label: 'Portals',
    icon: <PortalIcon color="currentColor" width={22} height={22} />,
    activeMatches: ['/hub-dashboard/portals/:id'],
  },
]
