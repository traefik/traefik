import { AriaTable, AriaTbody, AriaThead, AriaTr, Flex } from '@traefiklabs/faency'
import { orderBy } from 'lodash'
import { useContext, useEffect, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

import { SectionTitle } from './DetailsCard'

import AriaTableSkeleton from 'components/tables/AriaTableSkeleton'
import SortableTh from 'components/tables/SortableTh'
import { ToastContext } from 'contexts/toasts'
import { makeRowRender } from 'pages/http/HttpRouters'

type UsedByRoutersSectionProps = {
  data: Service.Details | Middleware.DetailsData
  protocol?: string
}

export const UsedByRoutersSkeleton = () => (
  <Flex gap={2} css={{ flexDirection: 'column', mt: '40px' }}>
    <SectionTitle title="Used by routers" />
    <AriaTableSkeleton columns={8} />
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
      <SectionTitle title="Used by Routers" />

      <AriaTable data-testid="routers-table" css={{ tableLayout: 'auto' }}>
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
