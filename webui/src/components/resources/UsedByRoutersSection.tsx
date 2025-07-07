import { AriaTable, AriaTbody, AriaTd, AriaTh, AriaThead, AriaTr, Box, Flex, styled } from '@traefiklabs/faency'
import { orderBy } from 'lodash'
import { useContext, useEffect, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

import { SectionHeader } from 'components/resources/DetailSections'
import SortableTh from 'components/tables/SortableTh'
import { ToastContext } from 'contexts/toasts'
import { MiddlewareDetailType, ServiceDetailType } from 'hooks/use-resource-detail'
import { makeRowRender } from 'pages/http/HttpRouters'

type UsedByRoutersSectionProps = {
  data: ServiceDetailType | MiddlewareDetailType
  protocol?: string
}

const SkeletonContent = styled(Box, {
  backgroundColor: '$slate5',
  height: '14px',
  minWidth: '50px',
  borderRadius: '4px',
  margin: '8px',
})

export const UsedByRoutersSkeleton = () => (
  <Flex css={{ flexDirection: 'column', mt: '40px' }}>
    <SectionHeader />
    <AriaTable>
      <AriaThead>
        <AriaTr>
          <AriaTh>
            <SkeletonContent />
          </AriaTh>
          <AriaTh>
            <SkeletonContent />
          </AriaTh>
          <AriaTh>
            <SkeletonContent />
          </AriaTh>
          <AriaTh>
            <SkeletonContent />
          </AriaTh>
          <AriaTh>
            <SkeletonContent />
          </AriaTh>
          <AriaTh>
            <SkeletonContent />
          </AriaTh>
        </AriaTr>
      </AriaThead>
      <AriaTbody>
        <AriaTr css={{ pointerEvents: 'none' }}>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
        </AriaTr>
        <AriaTr css={{ pointerEvents: 'none' }}>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
          <AriaTd>
            <SkeletonContent />
          </AriaTd>
        </AriaTr>
      </AriaTbody>
    </AriaTable>
  </Flex>
)

export const UsedByRoutersSection = ({ data, protocol = 'http' }: UsedByRoutersSectionProps) => {
  const renderRow = makeRowRender(protocol)
  const [searchParams] = useSearchParams()
  const { addToast } = useContext(ToastContext)

  const routersFound = useMemo(() => {
    let routers = data.routers?.filter((r) => !r.message)
    const direction = (searchParams.get('direction') as 'asc' | 'desc' | null) || 'asc'
    const sortBy = searchParams.get('sortBy') || 'name'
    if (sortBy) routers = orderBy(routers, [sortBy], [direction || 'asc'])
    return routers
  }, [data, searchParams])

  const routersNotFound = useMemo(() => data.routers?.filter((r) => !!r.message), [data])

  useEffect(() => {
    routersNotFound?.map((error) =>
      addToast({
        message: error.message,
        severity: 'error',
      }),
    )
  }, [addToast, routersNotFound])

  if (!routersFound || routersFound.length <= 0) {
    return null
  }

  return (
    <Flex css={{ flexDirection: 'column', mt: '$5' }}>
      <SectionHeader title="Used by Routers" />

      <AriaTable data-testid="routers-table">
        <AriaThead>
          <AriaTr>
            <SortableTh label="Status" css={{ width: '40px' }} isSortable sortByValue="status" />
            {protocol !== 'udp' ? <SortableTh css={{ width: '40px' }} label="TLS" /> : null}
            {protocol !== 'udp' ? <SortableTh label="Rule" isSortable sortByValue="rule" /> : null}
            <SortableTh label="Entrypoints" isSortable sortByValue="entryPoints" />
            <SortableTh label="Name" isSortable sortByValue="name" />
            <SortableTh label="Service" isSortable sortByValue="service" />
            <SortableTh label="Provider" css={{ width: '40px' }} isSortable sortByValue="provider" />
            <SortableTh label="Priority" isSortable sortByValue="priority" />
          </AriaTr>
        </AriaThead>
        <AriaTbody>{routersFound.map(renderRow)}</AriaTbody>
      </AriaTable>
    </Flex>
  )
}
