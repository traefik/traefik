import { ReactNode } from 'react'
import { FiHome } from 'react-icons/fi'
import { HiOutlineGlobe } from 'react-icons/hi'
import { TfiWorld } from 'react-icons/tfi'

export type NavRouteType = {
  path: string
  subPaths?: string[]
  label: string
  icon?: ReactNode
  subRoutes?: NavRouteType[]
}

const httpRoutes: NavRouteType[] = [
  {
    path: '/http/routers',
    subPaths: ['/http/routers/:name'],
    label: 'HTTP Routers',
  },
  {
    path: '/http/services',
    subPaths: ['/http/services/:name'],
    label: 'HTTP Services',
  },
  {
    path: '/http/middlewares',
    subPaths: ['/http/middlewares/:name'],
    label: 'HTTP Middlewares',
  },
]

const tcpRoutes: NavRouteType[] = [
  {
    path: '/tcp/routers',
    subPaths: ['/tcp/routers/:name'],
    label: 'TCP Routers',
  },
  {
    path: '/tcp/services',
    subPaths: ['/tcp/services/:name'],
    label: 'TCP Services',
  },
  {
    path: '/tcp/middlewares',
    subPaths: ['/tcp/middlewares/:name'],
    label: 'TCP Middlewares',
  },
]

const udpRoutes: NavRouteType[] = [
  {
    path: '/udp/routers',
    subPaths: ['/udp/routers/:name'],
    label: 'UDP Routers',
  },
  {
    path: '/udp/services',
    subPaths: ['/udp/services/:name'],
    label: 'UDP Services',
  },
]

const routes: NavRouteType[] = [
  { path: '/', label: 'Dashboard', icon: <FiHome color="currentColor" size={20} /> },
  {
    path: '/http/',
    label: 'HTTP',
    icon: <TfiWorld color="currentColor" size={20} />,
    subRoutes: httpRoutes,
  },
  {
    path: '/tcp/',
    label: 'TCP',
    icon: <HiOutlineGlobe color="currentColor" size={24} />,
    subRoutes: tcpRoutes,
  },
  {
    path: '/udp/',
    label: 'UDP',
    icon: <HiOutlineGlobe color="currentColor" size={24} />,
    subRoutes: udpRoutes,
  },
]

export default routes
