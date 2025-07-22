import useSWR from 'swr'

type TotalsResultItem = {
  routers: number
  services: number
  middlewares?: number
}

type TotalsResult = {
  http: TotalsResultItem
  tcp: TotalsResultItem
  udp: TotalsResultItem
}

const useTotals = (): TotalsResult => {
  const { data } = useSWR('/overview')

  return {
    http: {
      routers: data?.http?.routers?.total,
      services: data?.http?.services?.total,
      middlewares: data?.http?.middlewares?.total,
    },
    tcp: {
      routers: data?.tcp?.routers?.total,
      services: data?.tcp?.services?.total,
      middlewares: data?.tcp?.middlewares?.total,
    },
    udp: {
      routers: data?.udp?.routers?.total,
      services: data?.udp?.services?.total,
    },
  }
}

export default useTotals
