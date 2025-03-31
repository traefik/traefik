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
import useFetchWithPagination, { RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholder } from 'layout/EmptyPlaceholder'
import Page from 'layout/Page'
import { useEffect, useMemo, useState } from 'react'
import { FiArrowDown, FiArrowUp, FiShield } from 'react-icons/fi'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { NavigateFunction, useNavigate, useSearchParams } from 'react-router-dom'

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

export const HttpRoutersRender = ({
  error,
  isEmpty,
  isLoadingMore,
  isReachingEnd,
  loadMore,
  pageCount,
  pages,
  sortBy,
  direction,
  handleSort,
}) => {
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
            <Th onClick={() => handleSort('status')} style={{ cursor: 'pointer' }}>
              Status {sortBy === 'status' && (direction === 'asc' ? <FiArrowUp /> : <FiArrowDown />)}
            </Th>
            <Th>TLS</Th>
            <Th onClick={() => handleSort('rule')} style={{ cursor: 'pointer' }}>
              Rule {sortBy === 'rule' && (direction === 'asc' ? <FiArrowUp /> : <FiArrowDown />)}
            </Th>
            <Th onClick={() => handleSort('entrypoint')} style={{ cursor: 'pointer' }}>
              Entrypoints {sortBy === 'entrypoint' && (direction === 'asc' ? <FiArrowUp /> : <FiArrowDown />)}
            </Th>
            <Th onClick={() => handleSort('name')} style={{ cursor: 'pointer' }}>
              Name {sortBy === 'name' && (direction === 'asc' ? <FiArrowUp /> : <FiArrowDown />)}
            </Th>
            <Th onClick={() => handleSort('service')} style={{ cursor: 'pointer' }}>
              Service {sortBy === 'service' && (direction === 'asc' ? <FiArrowUp /> : <FiArrowDown />)}
            </Th>
            <Th onClick={() => handleSort('provider')} style={{ cursor: 'pointer' }}>
              Provider {sortBy === 'provider' && (direction === 'asc' ? <FiArrowUp /> : <FiArrowDown />)}
            </Th>
            <Th onClick={() => handleSort('priority')} style={{ cursor: 'pointer' }}>
              Priority {sortBy === 'priority' && (direction === 'asc' ? <FiArrowUp /> : <FiArrowDown />)}
            </Th>
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

  const [sortBy, setSortBy] = useState<string>('priority')
  const [direction, setDirection] = useState<'asc' | 'desc'>('desc')

  const handleSort = (column: string) => {
    setSortBy(column)
    setDirection((prevDirection) => (prevDirection === 'asc' ? 'desc' : 'asc'))
  }

  // const query = useMemo(() => searchParamsToState(searchParams), [searchParams.append("direction", direction)])
  const query = useMemo(
    () => ({ ...searchParamsToState(searchParams), direction, sortBy }),
    [searchParams, direction, sortBy],
  )
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
        sortBy={sortBy}
        direction={direction}
        handleSort={handleSort}
      />
    </Page>
  )
}
