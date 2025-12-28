import { AriaTable, AriaTbody, AriaTd, AriaTfoot, AriaThead, AriaTr, Flex } from '@traefiklabs/faency'
import { useMemo } from 'react'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { useSearchParams } from 'react-router-dom'

import { ScrollTopButton } from 'components/buttons/ScrollTopButton'
import { ProviderIconWithTooltip } from 'components/icons/providers'
import { Chips } from 'components/resources/DetailItemComponents'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { SpinnerLoader } from 'components/SpinnerLoader'
import ClickableRow from 'components/tables/ClickableRow'
import SortableTh from 'components/tables/SortableTh'
import { searchParamsToState, TableFilter } from 'components/tables/TableFilter'
import TooltipText from 'components/TooltipText'
import useFetchWithPagination, { pagesResponseInterface, RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholderTd } from 'layout/EmptyPlaceholder'
import PageTitle from 'layout/PageTitle'

export const makeRowRender = (): RenderRowType => {
  const UdpRoutersRenderRow = (row) => (
    <ClickableRow key={row.name} to={`/udp/routers/${row.name}`}>
      <AriaTd>
        <ResourceStatus status={row.status} />
      </AriaTd>
      <AriaTd>{row.entryPoints && row.entryPoints.length > 0 && <Chips items={row.entryPoints} />}</AriaTd>
      <AriaTd>
        <TooltipText text={row.name} isTruncated />
      </AriaTd>
      <AriaTd>
        <TooltipText text={row.service} isTruncated />
      </AriaTd>
      <AriaTd>
        <ProviderIconWithTooltip provider={row.provider} />
      </AriaTd>
      <AriaTd>
        <TooltipText text={row.priority} isTruncated />
      </AriaTd>
    </ClickableRow>
  )
  return UdpRoutersRenderRow
}

export const UdpRoutersRender = ({
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
            <SortableTh label="Entrypoints" isSortable sortByValue="entryPoints" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Service" isSortable sortByValue="service" />
            <SortableTh label="Provider" isSortable sortByValue="provider" />
            <SortableTh label="Priority" css={{ width: '60px' }} isSortable sortByValue="priority" />
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

export const UdpRouters = () => {
  const renderRow = makeRowRender()
  const [searchParams] = useSearchParams()

  const query = useMemo(() => searchParamsToState(searchParams), [searchParams])
  const { pages, pageCount, isLoadingMore, isReachingEnd, loadMore, error, isEmpty } = useFetchWithPagination(
    '/udp/routers',
    {
      listContextKey: JSON.stringify(query),
      renderRow,
      renderLoader: () => null,
      query,
    },
  )

  return (
    <>
      <PageTitle title="UDP Routers" />
      <TableFilter />
      <UdpRoutersRender
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
