import { Box, Flex, Td, Tfoot, Thead, Tr } from '@traefiklabs/faency'
import { useEffect, useMemo, useState } from 'react'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { useSearchParams } from 'react-router-dom'

import { AnimatedRow, AnimatedTable, AnimatedTBody } from 'components/AnimatedTable'
import { ProviderIcon } from 'components/resources/ProviderIcon'
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
      <AnimatedRow key={row.name} to={`/http/middlewares/${row.name}`}>
        <Td>
          <Tooltip label={row.status}>
            <Box css={{ width: '32px', height: '32px' }}>
              <ResourceStatus status={row.status} />
            </Box>
          </Tooltip>
        </Td>
        <Td>
          <TooltipText text={row.name} />
        </Td>
        <Td>
          <TooltipText text={middlewareType} />
        </Td>
        <Td>
          <Tooltip label={row.provider}>
            <Box css={{ width: '32px', height: '32px' }}>
              <ProviderIcon name={row.provider} />
            </Box>
          </Tooltip>
        </Td>
      </AnimatedRow>
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
  const [isMounted, setMounted] = useState(false)

  const [infiniteRef] = useInfiniteScroll({
    loading: isLoadingMore,
    hasNextPage: !isReachingEnd && !error,
    onLoadMore: loadMore,
  })

  useEffect(() => setMounted(true), [])

  return (
    <>
      <AnimatedTable>
        <Thead>
          <Tr>
            <SortableTh label="Status" css={{ width: '40px' }} isSortable sortByValue="status" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Type" isSortable sortByValue="type" />
            <SortableTh label="Provider" css={{ width: '40px' }} isSortable sortByValue="provider" />
          </Tr>
        </Thead>
        <AnimatedTBody pageCount={pageCount} isMounted={isMounted}>
          {pages}
        </AnimatedTBody>
        {(isEmpty || !!error) && (
          <Tfoot>
            <Tr>
              <Td colSpan={100}>
                <EmptyPlaceholder message={error ? 'Failed to fetch data' : 'No data available'} />
              </Td>
            </Tr>
          </Tfoot>
        )}
      </AnimatedTable>
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
