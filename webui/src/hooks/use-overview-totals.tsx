import { useMemo } from 'react'
import useSWR from 'swr'

export type UseTotalsProps = {
  protocol?: string
  enabled?: boolean
}

type TotalsResult = {
  routers: number
  services: number
  middlewares: number
}

const useTotals = ({ protocol, enabled = true }: UseTotalsProps): TotalsResult => {
  const { data } = useSWR(enabled ? '/overview' : null)

  return useMemo(
    () =>
      data && protocol
        ? {
            routers: data[protocol]?.routers?.total,
            services: data[protocol]?.services?.total,
            middlewares: data[protocol]?.middlewares?.total,
          }
        : ({} as TotalsResult),
    [data, protocol],
  )
}

export default useTotals
