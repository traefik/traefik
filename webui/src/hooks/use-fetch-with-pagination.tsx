import { AriaTd, AriaTr } from '@traefiklabs/faency'
import { stringify } from 'query-string'
import { ReactNode } from 'react'
import useSWRInfinite, { SWRInfiniteConfiguration } from 'swr/infinite'

import { fetchPage } from 'libs/fetch'

export type RenderRowType = (row: Record<string, unknown>) => ReactNode

export type pagesResponseInterface = {
  pages: ReactNode
  pageCount: number
  error?: Error | null
  isLoadingMore: boolean
  isReachingEnd: boolean
  isEmpty: boolean
  loadMore: () => void
}
type useFetchWithPaginationType = (
  path: string,
  opts: SWRInfiniteConfiguration & {
    rowsPerPage?: number
    renderRow: RenderRowType
    renderLoader?: () => ReactNode
    listContextKey?: string
    query?: Record<string, unknown>
  },
) => pagesResponseInterface

const useFetchWithPagination: useFetchWithPaginationType = (path, opts) => {
  const defaultLoadingFunction = () => (
    <AriaTr>
      <AriaTd>Loading...</AriaTd>
    </AriaTr>
  )
  const { rowsPerPage = 10, renderLoader = defaultLoadingFunction, renderRow, query } = opts

  const getKey = (
    pageIndex: number,
    previousPageData: { data?: unknown[]; nextPage?: number } | null,
  ): string | null => {
    if (previousPageData && (!previousPageData.data?.length || previousPageData.nextPage === 1)) return null

    return `${path}?${stringify({
      page: pageIndex + 1,
      per_page: rowsPerPage,
      ...query,
    })}`
  }

  const { data: res, error, size, setSize } = useSWRInfinite<{ data?: unknown[]; nextPage?: number }>(getKey, fetchPage)

  const isLoadingInitialData = !res && !error
  const isEmpty = !res?.[0]?.data || (Array.isArray(res?.[0]?.data) && res?.[0]?.data.length === 0)
  const isLoadingMore = isLoadingInitialData || (size > 0 && res && typeof res[size - 1] === 'undefined') || false
  const nextPage = res?.[size - 1]?.nextPage
  const isReachingEnd = !nextPage || nextPage === 1

  const loadMore = (): void => {
    if (!isLoadingMore) {
      setSize(size + 1)
    }
  }

  const data = res?.reduce((acc: unknown[], req) => {
    if (req.data) {
      acc.push(...req.data)
    }
    return acc
  }, [] as unknown[])

  let pages: ReactNode = null

  if (!error) {
    pages = !data ? renderLoader() : (data as Record<string, unknown>[]).map(renderRow)
  }

  return {
    pages,
    pageCount: size,
    isEmpty,
    error,
    isLoadingMore,
    isReachingEnd,
    loadMore,
  }
}

export default useFetchWithPagination
