import qs from 'query-string'
import { useMemo } from 'react'
import { useHref, useLocation, useSearchParams } from 'react-router'

import { capitalizeFirstLetter } from 'utils/string'

type UseGetUrlWithReturnTo = (href: string) => string

// path: pathname alone, used to detect a repeat. url: path + the page's own query params
// (returnTo excluded), used to rebuild a (possibly truncated) chain.
type ReturnToChainNode = { path: string; url: string; returnTo?: string }

// Hard cap on how many hops the returnTo chain can nest.
const MAX_RETURN_TO_CHAIN_DEPTH = 5

// Walks the nested returnTo chain encoded in a path+search string, one hop per level.
// Guards against a malformed/adversarial chain (a path repeating itself) with `seen`.
const decodeReturnToChain = (pathWithSearch: string): ReturnToChainNode[] => {
  const nodes: ReturnToChainNode[] = []
  const seen = new Set<string>()
  let current: string | undefined = pathWithSearch

  while (current) {
    const { url: path, query } = qs.parseUrl(current)
    if (seen.has(path)) break
    seen.add(path)

    const { returnTo, ...ownQuery } = query
    nodes.push({ path, url: qs.stringifyUrl({ url: path, query: ownQuery }), returnTo: returnTo as string })
    current = returnTo as string | undefined
  }

  return nodes
}

const buildReturnTo = (nodes: ReturnToChainNode[]): string | undefined =>
  nodes.reduceRight<string | undefined>(
    (returnTo, node) => (returnTo ? qs.stringifyUrl({ url: node.url, query: { returnTo } }) : node.url),
    undefined,
  )

export const useGetUrlWithReturnTo: UseGetUrlWithReturnTo = (href) => {
  const location = useLocation()
  const currentPath = location.pathname + location.search

  const url = useMemo(() => {
    if (!href) {
      return href
    }

    const chain = decodeReturnToChain(currentPath)

    // If the target is already part of the current returnTo chain, reuse it.
    const targetPath = href.split('?')[0]
    const existingNode = chain.find((node) => node.path === targetPath)

    // Once the chain would exceed MAX_RETURN_TO_CHAIN_DEPTH, drop the single oldest entry.
    const returnTo = existingNode
      ? existingNode.returnTo
      : chain.length > MAX_RETURN_TO_CHAIN_DEPTH
        ? buildReturnTo(chain.slice(0, MAX_RETURN_TO_CHAIN_DEPTH))
        : currentPath

    if (!returnTo) {
      return href
    }

    return qs.stringifyUrl({ url: href, query: { returnTo } })
  }, [currentPath, href])

  return url
}

export const useHrefWithReturnTo = (href: string): string => {
  const urlWithReturnTo = useGetUrlWithReturnTo(href)

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

type UseRouterReturnTo = () => {
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
