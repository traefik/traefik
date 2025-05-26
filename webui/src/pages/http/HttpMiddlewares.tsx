import { AriaTable, AriaTbody, AriaTd, AriaTfoot, AriaThead, AriaTr, Box, Flex } from '@traefiklabs/faency'
import { useMemo } from 'react'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { useSearchParams } from 'react-router-dom'

import ClickableRow from 'components/ClickableRow'
import ProviderIcon from 'components/icons/providers'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { ScrollTopButton } from 'components/ScrollTopButton'
import { SpinnerLoader } from 'components/SpinnerLoader'
import { searchParamsToState, TableFilter } from 'components/TableFilter'
import SortableTh from 'components/tables/SortableTh'
import Tooltip from 'components/Tooltip'
import TooltipText from 'components/TooltipText'
import useFetchWithPagination, { pagesResponseInterface, RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholder } from 'layout/EmptyPlaceholder'
import Page from 'layout/Page'
import { parseMiddlewareType } from 'libs/parsers'

export const makeRowRender = (): RenderRowType => {
  const HttpMiddlewaresRenderRow = (row) => {
    const middlewareType = parseMiddlewareType(row)

    return (
      <ClickableRow key={row.name} to={`/http/middlewares/${row.name}`}>
        <AriaTd>
          <Tooltip label={row.status}>
            <Box css={{ width: '32px', height: '32px' }}>
              <ResourceStatus status={row.status} />
            </Box>
          </Tooltip>
        </AriaTd>
        <AriaTd>
          <TooltipText text={row.name} />
        </AriaTd>
        <AriaTd>
          <TooltipText text={middlewareType} />
        </AriaTd>
        <AriaTd>
          <Tooltip label={row.provider}>
            <Box css={{ width: '32px', height: '32px' }}>
              <ProviderIcon name={row.provider} />
            </Box>
          </Tooltip>
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
            <SortableTh label="Status" css={{ width: '40px' }} isSortable sortByValue="status" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Type" isSortable sortByValue="type" />
            <SortableTh label="Provider" css={{ width: '75px' }} isSortable sortByValue="provider" />
          </AriaTr>
        </AriaThead>
        <AriaTbody>{pages}</AriaTbody>
        {(isEmpty || !!error) && (
          <AriaTfoot>
            <AriaTr>
              <AriaTd fullColSpan>
                <EmptyPlaceholder message={error ? 'Failed to fetch data' : 'No data available'} />
              </AriaTd>
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
    <Page title="HTTP Middlewares">
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
    </Page>
  )
}
