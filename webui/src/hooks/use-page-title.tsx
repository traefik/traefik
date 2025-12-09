import { useContext } from 'react'

import { VersionContext } from 'contexts/version'

const usePageTitle = (pageTitle?: string): string => {
  const { dashboardName } = useContext(VersionContext)

  return `${pageTitle ? `${pageTitle} - ` : ''}Traefik Proxy${dashboardName ? ` [${dashboardName}]` : ''}`
}

export default usePageTitle
