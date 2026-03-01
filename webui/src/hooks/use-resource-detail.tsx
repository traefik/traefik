import useSWR from 'swr'

import fetchMany from 'libs/fetchMany'

type ResourceDetailType = {
  data?: Resource.DetailsData
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
  const validMiddlewares = (middlewares as Middleware.Details[] | undefined)?.filter((mw) => !!mw.name)
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
