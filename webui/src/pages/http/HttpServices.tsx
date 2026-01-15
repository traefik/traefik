import { AriaTable, AriaTbody, AriaTd, AriaTfoot, AriaThead, AriaTr, Flex, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'
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
import PageTitle from 'layout/PageTitle'

export const makeRowRender = (): RenderRowType => {
  const HttpServicesRenderRow = (row) => (
    <ClickableRow key={row.name} to={`/http/services/${row.name}`}>
      <AriaTd>
        <ResourceStatus status={row.status} />
      </AriaTd>
      <AriaTd>
        <TooltipText text={row.name} />
      </AriaTd>
      <AriaTd>
        <TooltipText text={row.type} />
      </AriaTd>
      <AriaTd>
        <Text>{row.loadBalancer?.servers?.length || 0}</Text>
      </AriaTd>
      <AriaTd>
        <ProviderIconWithTooltip provider={row.provider} />
      </AriaTd>
    </ClickableRow>
  )
  return HttpServicesRenderRow
}

export const HttpServicesRender = ({
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
            <SortableTh label="Servers" isSortable sortByValue="servers" />
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

export const HttpServices = () => {
  const renderRow = makeRowRender()
  const [searchParams] = useSearchParams()

  const query = useMemo(() => searchParamsToState(searchParams), [searchParams])
  const { pages, pageCount, isLoadingMore, isReachingEnd, loadMore, error, isEmpty } = useFetchWithPagination(
    '/http/services',
    {
      listContextKey: JSON.stringify(query),
      renderRow,
      renderLoader: () => null,
      query,
    },
  )

  return (
    <>
      <PageTitle title="HTTP Services" />
      <TableFilter />
      <HttpServicesRender
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
