import { AriaTable, AriaTbody, AriaTd, AriaTfoot, AriaThead, AriaTr, Flex } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { Helmet } from 'react-helmet-async'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { useSearchParams } from 'react-router-dom'

import { ScrollTopButton } from 'components/buttons/ScrollTopButton'
import { ProviderIconWithTooltip } from 'components/icons/providers'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { SpinnerLoader } from 'components/SpinnerLoader'
import ClickableRow from 'components/tables/ClickableRow'
import SortableTh from 'components/tables/SortableTh'
import { searchParamsToState, TableFilter } from 'components/tables/TableFilter'
import TooltipText from 'components/TooltipText'
import useFetchWithPagination, { pagesResponseInterface, RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholderTd } from 'layout/EmptyPlaceholder'
import { parseMiddlewareType } from 'libs/parsers'

export const makeRowRender = (): RenderRowType => {
  const HttpMiddlewaresRenderRow = (row) => {
    const middlewareType = parseMiddlewareType(row)

    return (
      <ClickableRow key={row.name} to={`/http/middlewares/${row.name}`}>
        <AriaTd>
          <ResourceStatus status={row.status} />
        </AriaTd>
        <AriaTd>
          <TooltipText text={row.name} />
        </AriaTd>
        <AriaTd>
          <TooltipText text={middlewareType} />
        </AriaTd>
        <AriaTd>
          <ProviderIconWithTooltip provider={row.provider} />
        </AriaTd>
      </ClickableRow>
    )
  }
  return HttpMiddlewaresRenderRow
}

export const HttpMiddlewaresRender = ({
  error,
  isEmpty,
  isLoadingMore,
  isReachingEnd,
  loadMore,
  pageCount,
  pages,
}: pagesResponseInterface) => {
  const [infiniteRef] = useInfiniteScroll({
    loading: isLoadingMore,
    hasNextPage: !isReachingEnd && !error,
    onLoadMore: loadMore,
  })

  return (
    <>
      <AriaTable>
        <AriaThead>
          <AriaTr>
            <SortableTh label="Status" css={{ width: '36px' }} isSortable sortByValue="status" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Type" isSortable sortByValue="type" />
            <SortableTh label="Provider" css={{ width: '75px' }} isSortable sortByValue="provider" />
          </AriaTr>
        </AriaThead>
        <AriaTbody>{pages}</AriaTbody>
        {(isEmpty || !!error) && (
          <AriaTfoot>
            <AriaTr>
              <EmptyPlaceholderTd message={error ? 'Failed to fetch data' : 'No data available'} />
            </AriaTr>
          </AriaTfoot>
        )}
      </AriaTable>
      <Flex css={{ height: 60, alignItems: 'center', justifyContent: 'center' }} ref={infiniteRef}>
        {isLoadingMore ? <SpinnerLoader /> : isReachingEnd && pageCount > 1 && <ScrollTopButton />}
      </Flex>
    </>
  )
}

export const HttpMiddlewares = () => {
  const renderRow = makeRowRender()
  const [searchParams] = useSearchParams()

  const query = useMemo(() => searchParamsToState(searchParams), [searchParams])
  const { pages, pageCount, isLoadingMore, isReachingEnd, loadMore, error, isEmpty } = useFetchWithPagination(
    '/http/middlewares',
    {
      listContextKey: JSON.stringify(query),
      renderRow,
      renderLoader: () => null,
      query,
    },
  )

  return (
    <>
      <Helmet>
        <title>HTTP Middlewares - Traefik Proxy</title>
      </Helmet>
      <TableFilter />
      <HttpMiddlewaresRender
        error={error}
        isEmpty={isEmpty}
        isLoadingMore={isLoadingMore}
        isReachingEnd={isReachingEnd}
        loadMore={loadMore}
        pageCount={pageCount}
        pages={pages}
      />
    </>
  )
}
