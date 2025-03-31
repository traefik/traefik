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

  return data && protocol
    ? {
        routers: data[protocol]?.routers?.total,
        services: data[protocol]?.services?.total,
        middlewares: data[protocol]?.middlewares?.total,
      }
    : ({} as TotalsResult)
}

export default useTotals
