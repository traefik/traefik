namespace HubDemo {
  interface Route {
    path: string
    label: string
    icon: string
    contentPath: string
    dynamicSegments?: string[]
    activeMatches?: string[]
  }

  interface Manifest {
    version: string
    routes: Route[]
  }

  interface NavItem {
    path: string
    label: string
    icon: ReactNode
    activeMatches?: string[]
  }
}
