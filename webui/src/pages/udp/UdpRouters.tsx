import { Box, Flex, Td, Text, Tfoot, Th, Thead } from '@traefiklabs/faency'
import { AnimatedRow, AnimatedTable, AnimatedTBody } from 'components/AnimatedTable'
import { Tr } from 'components/FaencyOverrides'
import { Chips } from 'components/resources/DetailSections'
import { ProviderIcon } from 'components/resources/ProviderIcon'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { ScrollTopButton } from 'components/ScrollTopButton'
import { SpinnerLoader } from 'components/SpinnerLoader'
import { searchParamsToState, TableFilter } from 'components/TableFilter'
import Tooltip from 'components/Tooltip'
import useFetchWithPagination, { pagesResponseInterface, RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholder } from 'layout/EmptyPlaceholder'
import Page from 'layout/Page'
import { useEffect, useMemo, useState } from 'react'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { NavigateFunction, useNavigate, useSearchParams } from 'react-router-dom'

export const makeRowRender =
  (navigate: NavigateFunction): RenderRowType =>
  // eslint-disable-next-line react/display-name
  (row) =>
    (
      <AnimatedRow key={row.name} onClick={(): void => navigate(`/udp/routers/${row.name}`)}>
        <Td>
          <Tooltip label={row.status}>
            <Box css={{ width: '32px', height: '32px' }}>
              <ResourceStatus status={row.status} />
            </Box>
          </Tooltip>
        </Td>
        <Td>{row.entryPoints && row.entryPoints.length > 0 && <Chips items={row.entryPoints} />}</Td>
        <Td>
          <Tooltip label={row.name} action="copy">
            <Text css={{ wordBreak: 'break-word' }}>{row.name}</Text>
          </Tooltip>
        </Td>
        <Td>
          <Tooltip label={row.service} action="copy">
            <Text css={{ wordBreak: 'break-word' }}>{row.service}</Text>
          </Tooltip>
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

export const UdpRoutersRender = ({
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
            <Th style={{ width: '33%' }}>Entrypoints</Th>
            <Th style={{ width: '33%' }}>Name</Th>
            <Th style={{ width: '33%' }}>Service</Th>
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

export const UdpRouters = () => {
  const navigate = useNavigate()
  const renderRow = makeRowRender(navigate)
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
    <Page title="UDP Routers">
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
    </Page>
  )
}
