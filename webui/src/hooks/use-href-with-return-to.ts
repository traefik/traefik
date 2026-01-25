import qs from 'query-string'
import { useMemo } from 'react'
import { useHref, useLocation, useSearchParams } from 'react-router-dom'

import { capitalizeFirstLetter } from '../utils/string'

type UseGetUrlWithReturnTo = (href: string, initialReturnTo?: string) => string

export const useGetUrlWithReturnTo: UseGetUrlWithReturnTo = (href, initialReturnTo) => {
  const location = useLocation()
  const currentPath = location.pathname + location.search

  const url = useMemo(() => {
    if (href) {
      return qs.stringifyUrl({ url: href, query: { returnTo: initialReturnTo ?? currentPath } })
    }
    return href
  }, [currentPath, href, initialReturnTo])

  return url
}

export const useHrefWithReturnTo = (href: string, returnTo?: string): string => {
  const urlWithReturnTo = useGetUrlWithReturnTo(href, returnTo)

  return useHref(urlWithReturnTo)
}

const RETURN_TO_LABEL_OVERRIDES_SINGULAR: Record<string, Record<string, string>> = {
  http: {
    routers: 'HTTP router',
    services: 'HTTP service',
    middlewares: 'HTTP middleware',
  },
  tcp: {
    routers: 'TCP router',
    services: 'TCP service',
    middlewares: 'TCP middleware',
  },
  udp: {
    routers: 'UDP router',
    services: 'UDP service',
  },
}

const RETURN_TO_LABEL_OVERRIDES_PLURAL: Record<string, Record<string, string>> = {
  http: {
    routers: 'HTTP routers',
    services: 'HTTP services',
    middlewares: 'HTTP middlewares',
  },
  tcp: {
    routers: 'TCP routers',
    services: 'TCP services',
    middlewares: 'TCP middlewares',
  },
  udp: {
    routers: 'UDP routers',
    services: 'UDP services',
  },
}

type UseRouterReturnTo = (initialReturnTo?: string) => {
  returnTo: string | null
  returnToLabel: string | null
}

const getCleanPath = (path: string) => {
  if (!path) return ''
  return path.split('?')[0]
}

export const useRouterReturnTo: UseRouterReturnTo = () => {
  const [searchParams] = useSearchParams()

  const returnTo = useMemo(() => {
    const queryReturnTo = searchParams.get('returnTo')
    return queryReturnTo || null
  }, [searchParams])

  const returnToHref = useHref(returnTo || '')

  const returnToLabel = useMemo(() => {
    if (!returnTo) {
      return null
    }

    const returnToArr = returnTo.split('/')

    const [, path, subpath, id] = returnToArr

    // Strip query params from path, if any
    const cleanPath = getCleanPath(path)
    const cleanSubpath = getCleanPath(subpath)

    // Malformed returnTo (e.g., just '/' or empty path)
    if (!cleanPath) {
      return 'Back'
    }

    const fallbackLabel = `${capitalizeFirstLetter(cleanPath)}${cleanSubpath ? ` ${cleanSubpath}` : ''}`

    const labelArray = id ? RETURN_TO_LABEL_OVERRIDES_SINGULAR : RETURN_TO_LABEL_OVERRIDES_PLURAL

    const labelOverride =
      labelArray[cleanPath]?.[cleanSubpath] ??
      (typeof labelArray[cleanPath] === 'string' ? labelArray[cleanPath] : fallbackLabel)

    return capitalizeFirstLetter(labelOverride)
  }, [returnTo])

  return useMemo(
    () => ({
      returnTo: returnTo ? returnToHref : null,
      returnToLabel,
    }),
    [returnTo, returnToHref, returnToLabel],
  )
}
