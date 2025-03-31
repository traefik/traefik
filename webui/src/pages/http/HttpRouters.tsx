import { Box, Flex, Td, Text, Tfoot, Thead } from '@traefiklabs/faency'
import { useEffect, useMemo, useState } from 'react'
import { FiShield } from 'react-icons/fi'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { NavigateFunction, useNavigate, useSearchParams } from 'react-router-dom'

import { AnimatedRow, AnimatedTable, AnimatedTBody } from 'components/AnimatedTable'
import { Tr } from 'components/FaencyOverrides'
import { Chips } from 'components/resources/DetailSections'
import { ProviderIcon } from 'components/resources/ProviderIcon'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { ScrollTopButton } from 'components/ScrollTopButton'
import { SpinnerLoader } from 'components/SpinnerLoader'
import { searchParamsToState, TableFilter } from 'components/TableFilter'
import SortableTh from 'components/tables/SortableTh'
import Tooltip from 'components/Tooltip'
import useFetchWithPagination, { RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholder } from 'layout/EmptyPlaceholder'
import Page from 'layout/Page'

export const makeRowRender = (navigate: NavigateFunction, protocol = 'http'): RenderRowType => {
  const HttpRoutersRenderRow = (row) => (
    <AnimatedRow key={row.name} onClick={(): void => navigate(`/${protocol}/routers/${row.name}`)}>
      <Td>
        <Tooltip label={row.status}>
          <Box css={{ width: '32px', height: '32px' }}>
            <ResourceStatus status={row.status} />
          </Box>
        </Tooltip>
      </Td>
      {protocol !== 'udp' && (
        <>
          <Td>
            {row.tls && (
              <Tooltip label="TLS ON">
                <Box css={{ width: 24, height: 24 }} data-testid="tls-on">
                  <FiShield color="#008000" fill="#008000" size={24} />
                </Box>
              </Tooltip>
            )}
          </Td>
          <Td>
            <Tooltip label={row.rule} action="copy">
              <Text css={{ wordBreak: 'break-word' }}>{row.rule}</Text>
            </Tooltip>
          </Td>
        </>
      )}
      <Td>{row.using && row.using.length > 0 && <Chips items={row.using} />}</Td>
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
      <Td>
        <Tooltip label={row.priority} action="copy">
          <Text css={{ wordBreak: 'break-word' }}>{row.priority}</Text>
        </Tooltip>
      </Td>
    </AnimatedRow>
  )
  return HttpRoutersRenderRow
}

export const HttpRoutersRender = ({ error, isEmpty, isLoadingMore, isReachingEnd, loadMore, pageCount, pages }) => {
  const [isMounted, setMounted] = useState(false)

  const [infiniteRef] = useInfiniteScroll({
    loading: isLoadingMore,
    hasNextPage: !isReachingEnd && !error,
    onLoadMore: loadMore,
  })

  useEffect(() => setMounted(true), [])

  return (
    <>
      <AnimatedTable isMounted={isMounted}>
        <Thead>
          <Tr>
            <SortableTh label="Status" css={{ width: '40px' }} isSortable sortByValue="status" />
            <SortableTh label="TLS" />
            <SortableTh label="Rule" isSortable sortByValue="rule" />
            <SortableTh label="Entrypoints" isSortable sortByValue="entrypoint" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Service" isSortable sortByValue="service" />
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

export const HttpRouters = () => {
  const navigate = useNavigate()
  const renderRow = makeRowRender(navigate)
  const [searchParams] = useSearchParams()

  const query = useMemo(() => searchParamsToState(searchParams), [searchParams])
  const { pages, pageCount, isLoadingMore, isReachingEnd, loadMore, error, isEmpty } = useFetchWithPagination(
    '/http/routers',
    {
      listContextKey: JSON.stringify(query),
      renderRow,
      renderLoader: () => null,
      query,
    },
  )

  return (
    <Page title="HTTP Routers">
      <TableFilter />
      <HttpRoutersRender
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
