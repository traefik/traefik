import { Box, Flex, Td, Tfoot, Thead, Tr } from '@traefiklabs/faency'
import { useEffect, useMemo, useState } from 'react'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { NavigateFunction, useNavigate, useSearchParams } from 'react-router-dom'

import { AnimatedRow, AnimatedTable, AnimatedTBody } from 'components/AnimatedTable'
import { Chips } from 'components/resources/DetailSections'
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

export const makeRowRender = (navigate: NavigateFunction): RenderRowType => {
  const TcpRoutersRenderRow = (row) => (
    <AnimatedRow key={row.name} onClick={(): void => navigate(`/tcp/routers/${row.name}`)}>
      <Td>
        <Tooltip label={row.status}>
          <Box css={{ width: '32px', height: '32px' }}>
            <ResourceStatus status={row.status} />
          </Box>
        </Tooltip>
      </Td>
      <Td>
        <TooltipText text={row.rule} isTruncated variant="wide" />
      </Td>
      <Td>{row.entryPoints && row.entryPoints.length > 0 && <Chips items={row.entryPoints} />}</Td>
      <Td>
        <TooltipText text={row.name} isTruncated />
      </Td>
      <Td>
        <Tooltip label={row.provider}>
          <Box css={{ width: '32px', height: '32px' }}>
            <ProviderIcon name={row.provider} />
          </Box>
        </Tooltip>
      </Td>
      <Td>
        <TooltipText text={row.priority} isTruncated variant="short" />
      </Td>
    </AnimatedRow>
  )
  return TcpRoutersRenderRow
}

export const TcpRoutersRender = ({
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
            <SortableTh label="Rule" isSortable sortByValue="rule" />
            <SortableTh label="Entrypoints" isSortable sortByValue="entryPoints" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Provider" css={{ width: '40px' }} isSortable sortByValue="provider" />
            <SortableTh label="Priority" css={{ width: '64px' }} isSortable sortByValue="priority" />
          </Tr>
        </Thead>
        <AnimatedTBody pageCount={pageCount} isMounted={isMounted}>
          {pages}
        </AnimatedTBody>
        {(isEmpty || !!error) && (
          <Tfoot>
            <Tr>
              <td colSpan={100}>
                <EmptyPlaceholder message={error ? 'Failed to fetch data' : 'No data available'} />
              </td>
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

export const TcpRouters = () => {
  const navigate = useNavigate()
  const renderRow = makeRowRender(navigate)
  const [searchParams] = useSearchParams()

  const query = useMemo(() => searchParamsToState(searchParams), [searchParams])
  const { pages, pageCount, isLoadingMore, isReachingEnd, loadMore, error, isEmpty } = useFetchWithPagination(
    '/tcp/routers',
    {
      listContextKey: JSON.stringify(query),
      renderRow,
      renderLoader: () => null,
      query,
    },
  )

  return (
    <Page title="TCP Routers">
      <TableFilter />
      <TcpRoutersRender
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
