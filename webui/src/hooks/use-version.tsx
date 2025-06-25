import { useMemo } from 'react'
import useSWR from 'swr'

export default function useVersion() {
  const { data: version } = useSWR('/version')

  const showHubButton = useMemo(() => {
    if (!version) return false
    return !version?.disableDashboardAd
  }, [version])

  return { showHubButton, version }
}
