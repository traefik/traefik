import { createContext, ReactNode, useEffect, useState } from 'react'

import { BASE_PATH } from 'libs/utils'

type VersionContextProps = {
  showHubButton: boolean
  version: string
  showDemoSection: boolean
}

export const VersionContext = createContext<VersionContextProps>({
  showHubButton: false,
  version: '',
  showDemoSection: false,
})

type VersionProviderProps = {
  children: ReactNode
}

export const VersionProvider = ({ children }: VersionProviderProps) => {
  const [showHubButton, setShowHubButton] = useState(false)
  const [showDemoSection, setShowDemoSection] = useState(false)
  const [version, setVersion] = useState('')

  useEffect(() => {
    const fetchVersion = async () => {
      try {
        const response = await fetch(`${BASE_PATH}/version`)
        if (!response.ok) {
          throw new Error(`Network error: ${response.status}`)
        }
        const data: API.Version = await response.json()
        setShowHubButton(!data.disableDashboardAd)
        setShowDemoSection(!data.disableDashboardDemo)
        setVersion(data.Version)
      } catch (err) {
        console.error(err)
      }
    }

    fetchVersion()
  }, [])

  return <VersionContext.Provider value={{ showHubButton, version, showDemoSection }}>{children}</VersionContext.Provider>
}
