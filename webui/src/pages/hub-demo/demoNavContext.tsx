import { createContext } from 'react'
import { RouteObject } from 'react-router-dom'

import { useHubDemo } from './use-hub-demo'

export const HubDemoContext = createContext<{
  routes: RouteObject[] | null
  navigationItems: HubDemo.NavItem[] | null
}>({ routes: null, navigationItems: null })

export const HubDemoProvider = ({ basePath, children }) => {
  const { routes, navigationItems } = useHubDemo(basePath)

  return <HubDemoContext.Provider value={{ routes, navigationItems }}>{children}</HubDemoContext.Provider>
}
