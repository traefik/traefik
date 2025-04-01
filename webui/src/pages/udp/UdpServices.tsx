import { Box, Flex, Td, Text, Tfoot, Th, Thead } from '@traefiklabs/faency'
import { useEffect, useMemo, useState } from 'react'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { NavigateFunction, useNavigate, useSearchParams } from 'react-router-dom'

import { AnimatedRow, AnimatedTable, AnimatedTBody } from 'components/AnimatedTable'
import { Tr } from 'components/FaencyOverrides'
import { ProviderIcon } from 'components/resources/ProviderIcon'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { ScrollTopButton } from 'components/ScrollTopButton'
import { SpinnerLoader } from 'components/SpinnerLoader'
import { searchParamsToState, TableFilter } from 'components/TableFilter'
import Tooltip from 'components/Tooltip'
import useFetchWithPagination, { pagesResponseInterface, RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholder } from 'layout/EmptyPlaceholder'
import Page from 'layout/Page'

export const makeRowRender = (navigate: NavigateFunction): RenderRowType => {
  const UdpServicesRenderRow = (row) => (
    <AnimatedRow key={row.name} onClick={(): void => navigate(`/udp/services/${row.name}`)}>
      <Td>
        <Tooltip label={row.status}>
          <Box css={{ width: '32px', height: '32px' }}>
            <ResourceStatus status={row.status} />
          </Box>
        </Tooltip>
      </Td>
      <Td>
        <Tooltip label={row.name} action="copy">
          <Text>{row.name}</Text>
        </Tooltip>
      </Td>
      <Td>
        <Tooltip label={row.type} action="copy">
          <Text>{row.type}</Text>
        </Tooltip>
      </Td>
      <Td style={{ textAlign: 'right' }}>
        <Text>{row.loadBalancer?.servers?.length || 0}</Text>
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
        <Thead>
          <Tr>
            <Th style={{ width: '72px' }}>Status</Th>
            <Th style={{ width: '50%' }}>Name</Th>
            <Th style={{ width: '50%' }}>Type</Th>
            <Th style={{ textAlign: 'right', width: '80px' }}>Servers</Th>
            <Th style={{ textAlign: 'right', width: '80px' }}>Provider</Th>
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

export const UdpServices = () => {
  const navigate = useNavigate()
  const renderRow = makeRowRender(navigate)
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
