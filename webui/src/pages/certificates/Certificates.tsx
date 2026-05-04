import { AriaTable, AriaTbody, AriaTd, AriaTfoot, AriaThead, AriaTr, Flex, Text } from '@traefik-labs/faency'
import { useMemo } from 'react'
import useInfiniteScroll from 'react-infinite-scroll-hook'
import { useSearchParams } from 'react-router-dom'

import { ScrollTopButton } from 'components/buttons/ScrollTopButton'
import CertExpiryBadge from 'components/certificates/CertExpiryBadge'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { SpinnerLoader } from 'components/SpinnerLoader'
import ClickableRow from 'components/tables/ClickableRow'
import SortableTh from 'components/tables/SortableTh'
import { searchParamsToState, TableFilter } from 'components/tables/TableFilter'
import TooltipText from 'components/TooltipText'
import { computeDaysLeft } from 'hooks/use-certificates'
import useFetchWithPagination, { pagesResponseInterface, RenderRowType } from 'hooks/use-fetch-with-pagination'
import { EmptyPlaceholderTd } from 'layout/EmptyPlaceholder'
import PageTitle from 'layout/PageTitle'

export const CertificateRenderRow: RenderRowType = (row: unknown) => {
  const cert = row as Certificate.Raw
  const daysLeft = computeDaysLeft(cert.notAfter)
  const validUntil = new Date(cert.notAfter).toLocaleDateString()

  return (
    <ClickableRow key={cert.name} to={`/certificates/${cert.name}`}>
      <AriaTd>
        <ResourceStatus status={cert.status} />
      </AriaTd>
      <AriaTd>
        <TooltipText text={cert.commonName || '-'} />
      </AriaTd>
      <AriaTd css={{ maxWidth: '240px' }}>
        <Text css={{ wordBreak: 'break-word', whiteSpace: 'normal' }}>
          {cert.sans?.length > 0 ? cert.sans.join(', ') : '-'}
        </Text>
      </AriaTd>
      <AriaTd>
        <TooltipText text={cert.issuerOrg || cert.issuerCN || 'Unknown'} />
      </AriaTd>
      <AriaTd>
        <Text>{validUntil}</Text>
      </AriaTd>
      <AriaTd>
        <CertExpiryBadge daysLeft={daysLeft} size="small" />
      </AriaTd>
    </ClickableRow>
  )
}

export const CertificatesRender = ({
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
            <SortableTh label="Status" isSortable sortByValue="status" css={{ width: '36px' }} />
            <SortableTh label="Common Name" isSortable sortByValue="cn" />
            <SortableTh label="SANs" css={{ maxWidth: '240px' }} />
            <SortableTh label="Issuer" isSortable sortByValue="issuer" />
            <SortableTh label="Valid Until" isSortable sortByValue="validUntil" css={{ width: '100px' }} />
            <SortableTh label="Expiry" css={{ width: '100px' }} />
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

export const Certificates = () => {
  const [searchParams] = useSearchParams()

  const query = useMemo(() => searchParamsToState(searchParams), [searchParams])
  const { pages, pageCount, isLoadingMore, isReachingEnd, loadMore, error, isEmpty } = useFetchWithPagination(
    '/certificates',
    {
      listContextKey: JSON.stringify(query),
      renderRow: CertificateRenderRow,
      renderLoader: () => null,
      query,
    },
  )

  return (
    <>
      <PageTitle title="Certificates" />
      <TableFilter />
      <CertificatesRender
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
