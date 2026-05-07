import { createContext, ReactNode, useEffect, useState } from 'react'

import { BASE_PATH } from 'libs/utils'

export type DashboardNamePosition = 'side' | 'below'

type VersionContextProps = {
  showHubButton: boolean
  version: string
  dashboardName: string
  dashboardNamePosition: DashboardNamePosition
}

export const VersionContext = createContext<VersionContextProps>({
  showHubButton: false,
  version: '',
  dashboardName: '',
  dashboardNamePosition: 'side',
})

type VersionProviderProps = {
  children: ReactNode
}

const normalizePosition = (raw: string | undefined): DashboardNamePosition => {
  return raw === 'below' ? 'below' : 'side'
}

export const VersionProvider = ({ children }: VersionProviderProps) => {
  const [showHubButton, setShowHubButton] = useState(false)
  const [version, setVersion] = useState('')
  const [dashboardName, setDashboardName] = useState('')
  const [dashboardNamePosition, setDashboardNamePosition] = useState<DashboardNamePosition>('side')

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
        setDashboardNamePosition(normalizePosition(data.dashboardNamePosition))
      } catch (err) {
        console.error(err)
      }
    }

    fetchVersion()
  }, [])

  return (
    <VersionContext.Provider
      value={{ showHubButton, version, dashboardName, dashboardNamePosition }}
    >
      {children}
    </VersionContext.Provider>
  )
}
