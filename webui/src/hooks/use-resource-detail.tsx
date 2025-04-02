import useSWR from 'swr'

import fetchMany from 'libs/fetchMany'

export type EntryPoint = {
  name: string
  address: string
  message?: string
}

type JSONObject = {
  [x: string]: string | number
}
export type ValuesMapType = {
  [key: string]: string | number | JSONObject
}

export type MiddlewareProps = {
  [prop: string]: ValuesMapType
}

export type Middleware = {
  name: string
  status: 'enabled' | 'disabled' | 'warning'
  provider: string
  type?: string
  plugin?: Record<string, unknown>
  error?: string[]
  routers?: string[]
  usedBy?: string[]
} & MiddlewareProps

type Router = {
  name: string
  service?: string
  status: 'enabled' | 'disabled' | 'warning'
  rule?: string
  priority?: number
  provider: string
  tls?: {
    options: string
    certResolver: string
    domains: TlsDomain[]
    passthrough: boolean
  }
  error?: string[]
  entryPoints?: string[]
  message?: string
}

type TlsDomain = {
  main: string
  sans: string[]
}

export type RouterDetailType = Router & {
  middlewares?: Middleware[]
  hasValidMiddlewares?: boolean
  entryPointsData?: EntryPoint[]
  using?: string[]
}

type Mirror = {
  name: string
  percent: number
}

export type ServiceDetailType = {
  name: string
  status: 'enabled' | 'disabled' | 'warning'
  provider: string
  type: string
  usedBy?: string[]
  routers?: Router[]
  serverStatus?: {
    [server: string]: string
  }
  mirroring?: {
    service: string
    mirrors?: Mirror[]
  }
  loadBalancer?: {
    servers?: { url: string }[]
    passHostHeader?: boolean
    terminationDelay?: number
    healthCheck?: {
      scheme: string
      path: string
      port: number
      interval: string
      timeout: string
      hostname: string
      headers?: {
        [header: string]: string
      }
    }
  }
  weighted?: {
    services?: {
      name: string
      weight: number
    }[]
  }
}

export type MiddlewareDetailType = Middleware & {
  routers?: Router[]
}

export type ResourceDetailDataType = RouterDetailType & ServiceDetailType & MiddlewareDetailType

type ResourceDetailType = {
  data?: ResourceDetailDataType
  error?: Error
}

export const useResourceDetail = (name: string, resource: string, protocol = 'http'): ResourceDetailType => {
  const { data: routeDetail, error } = useSWR(`/${protocol}/${resource}/${name}`)
  const { data: entryPoints, error: entryPointsError } = useSWR(() => ['/entrypoints/', routeDetail.using], fetchMany)
  const { data: middlewares, error: middlewaresError } = useSWR(
    () => [`/${protocol}/middlewares/`, routeDetail.middlewares],
    fetchMany,
  )
  const { data: routers, error: routersError } = useSWR(() => [`/${protocol}/routers/`, routeDetail.usedBy], fetchMany)

  if (!routeDetail) {
    return { error }
  }

  const firstError = error || entryPointsError || middlewaresError || routersError
  const validMiddlewares = (middlewares as Middleware[] | undefined)?.filter((mw) => !!mw.name)
  const hasMiddlewares = validMiddlewares
    ? validMiddlewares.length > 0
    : routeDetail.middlewares && routeDetail.middlewares.length > 0

  if (resource === 'routers') {
    return {
      data: {
        name: routeDetail.name,
        service: routeDetail.service,
        status: routeDetail.status,
        provider: routeDetail.provider,
        rule: routeDetail.rule,
        tls: routeDetail.tls,
        error: routeDetail.error,
        middlewares: validMiddlewares,
        hasValidMiddlewares: hasMiddlewares,
        entryPointsData: entryPoints,
        using: routeDetail.using,
      },
      error: firstError,
    } as ResourceDetailType
  }

  if (resource === 'middlewares') {
    return {
      data: {
        ...routeDetail,
        routers,
      },
      error: firstError,
    } as ResourceDetailType
  }

  return {
    data: {
      name: routeDetail.name,
      status: routeDetail.status,
      provider: routeDetail.provider,
      type: routeDetail.type,
      loadBalancer: routeDetail.loadBalancer,
      mirroring: routeDetail.mirroring,
      serverStatus: routeDetail.serverStatus,
      usedBy: routeDetail.usedBy,
      weighted: routeDetail.weighted,
      routers,
    },
    error: firstError,
  } as ResourceDetailType
}
