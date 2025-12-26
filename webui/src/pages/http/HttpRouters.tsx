import { AriaTable, AriaTbody, AriaTd, AriaTfoot, AriaThead, AriaTr, Box, Flex } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { Helmet } from 'react-helmet-async'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { useSearchParams } from 'react-router-dom'

import { ScrollTopButton } from 'components/buttons/ScrollTopButton'
import { ProviderIconWithTooltip } from 'components/icons/providers'
import { Chips } from 'components/resources/DetailItemComponents'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import TlsIcon from 'components/routers/TlsIcon'
import { SpinnerLoader } from 'components/SpinnerLoader'
import ClickableRow from 'components/tables/ClickableRow'
import SortableTh from 'components/tables/SortableTh'
import { searchParamsToState, TableFilter } from 'components/tables/TableFilter'
import Tooltip from 'components/Tooltip'
import TooltipText from 'components/TooltipText'
import useFetchWithPagination, { pagesResponseInterface, RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholderTd } from 'layout/EmptyPlaceholder'

export const makeRowRender = (protocol = 'http'): RenderRowType => {
  const HttpRoutersRenderRow = (row) => (
    <ClickableRow key={row.name} to={`/${protocol}/routers/${row.name}`}>
      <AriaTd>
        <ResourceStatus status={row.status} />
      </AriaTd>
      {protocol !== 'udp' && (
        <>
          <AriaTd>
            {row.tls && (
              <Tooltip label="TLS ON">
                <Box css={{ width: 20, height: 20 }} data-testid="tls-on">
                  <TlsIcon />
                </Box>
              </Tooltip>
            )}
          </AriaTd>
          <AriaTd>
            <TooltipText
              text={row.rule}
              css={{
                display: '-webkit-box',
                '-webkit-line-clamp': 2,
                '-webkit-box-orient': 'vertical',
                overflow: 'hidden',
                wordBreak: 'break-word',
                maxWidth: '100%',
                lineHeight: 1.3,
              }}
            />
          </AriaTd>
        </>
      )}
      <AriaTd>{row.using && row.using.length > 0 && <Chips items={row.using} />}</AriaTd>
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
            <SortableTh label="TLS" css={{ width: '24px' }} />
            <SortableTh label="Rule" isSortable sortByValue="rule" />
            <SortableTh label="Entrypoints" isSortable sortByValue="entryPoints" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Service" isSortable sortByValue="service" />
            <SortableTh label="Provider" css={{ width: '40px' }} isSortable sortByValue="provider" />
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

export const HttpRouters = () => {
  const renderRow = makeRowRender()
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
    <>
      <Helmet>
        <title>HTTP Routers - Traefik Proxy</title>
      </Helmet>
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
    </>
  )
}
