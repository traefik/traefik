import { createContext, ReactNode, useEffect, useState } from 'react'

import { BASE_PATH } from 'libs/utils'

type VersionContextProps = {
  showHubButton: boolean
  version: string
  dashboardName: string
}

export const VersionContext = createContext<VersionContextProps>({
  showHubButton: false,
  version: '',
  dashboardName: '',
})

type VersionProviderProps = {
  children: ReactNode
}

export const VersionProvider = ({ children }: VersionProviderProps) => {
  const [showHubButton, setShowHubButton] = useState(false)
  const [version, setVersion] = useState('')
  const [dashboardName, setDashboardName] = useState('')

  useEffect(() => {
    const fetchVersion = async () => {
      try {
        const response = await fetch(`${BASE_PATH}/version`)
        if (!response.ok) {
          throw new Error(`Network error: ${response.status}`)
        }
        const data: API.Version = await response.json()
        setShowHubButton(!data.disableDashboardAd)
        setVersion(data.Version)
        setDashboardName(data.dashboardName || '')
      } catch (err) {
        console.error(err)
      }
    }

    fetchVersion()
  }, [])

  return <VersionContext.Provider value={{ showHubButton, version, dashboardName }}>{children}</VersionContext.Provider>
}
