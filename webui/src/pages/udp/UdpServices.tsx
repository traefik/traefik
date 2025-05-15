import { AriaTd, AriaTfoot, AriaThead, AriaTr, Box, Flex, Text } from '@traefiklabs/faency'
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

export const makeRowRender = (): RenderRowType => {
  const UdpServicesRenderRow = (row) => (
    <AnimatedRow key={row.name} to={`/udp/services/${row.name}`}>
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
        <TooltipText text={row.type} />
      </AriaTd>
      <AriaTd style={{ textAlign: 'right' }}>
        <Text>{row.loadBalancer?.servers?.length || 0}</Text>
      </AriaTd>
      <AriaTd>
        <Tooltip label={row.provider}>
          <Box css={{ width: '32px', height: '32px' }}>
            <ProviderIcon name={row.provider} />
          </Box>
        </Tooltip>
      </AriaTd>
    </AnimatedRow>
  )
  return UdpServicesRenderRow
}

export const UdpServicesRender = ({
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
        <AriaThead>
          <AriaTr>
            <SortableTh label="Status" css={{ width: '40px' }} isSortable sortByValue="status" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Type" isSortable sortByValue="type" />
            <SortableTh label="Servers" css={{ width: '40px' }} isSortable sortByValue="servers" />
            <SortableTh label="Provider" css={{ width: '40px' }} isSortable sortByValue="provider" />
          </AriaTr>
        </AriaThead>
        <AnimatedTBody pageCount={pageCount} isMounted={isMounted}>
          {pages}
        </AnimatedTBody>
        {(isEmpty || !!error) && (
          <AriaTfoot>
            <AriaTr>
              <AriaTd fullColSpan>
                <EmptyPlaceholder message={error ? 'Failed to fetch data' : 'No data available'} />
              </AriaTd>
            </AriaTr>
          </AriaTfoot>
        )}
      </AnimatedTable>
      <Flex css={{ height: 60, alignItems: 'center', justifyContent: 'center' }} ref={infiniteRef}>
        {isLoadingMore ? <SpinnerLoader /> : isReachingEnd && pageCount > 1 && <ScrollTopButton />}
      </Flex>
    </>
  )
}

export const UdpServices = () => {
  const renderRow = makeRowRender()
  const [searchParams] = useSearchParams()

  const query = useMemo(() => searchParamsToState(searchParams), [searchParams])
  const { pages, pageCount, isLoadingMore, isReachingEnd, loadMore, error, isEmpty } = useFetchWithPagination(
    '/udp/services',
    {
      listContextKey: JSON.stringify(query),
      renderRow,
      renderLoader: () => null,
      query,
    },
  )

  return (
    <Page title="UDP Services">
      <TableFilter />
      <UdpServicesRender
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
