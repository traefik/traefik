import { Flex } from '@traefiklabs/faency'
import { orderBy } from 'lodash'
import { useContext, useEffect, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

import { SectionTitle } from './DetailsCard'

import AriaTableSkeleton from 'components/tables/AriaTableSkeleton'
import PaginatedTable from 'components/tables/PaginatedTable'
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

  const columns = useMemo((): Array<{
    key: keyof Router.DetailsData
    header: string
    sortable?: boolean
    width?: string
  }> => {
    return [
      { key: 'status', header: 'Status', sortable: true, width: '36px' },
      ...(protocol !== 'udp' ? [{ key: 'tls' as keyof Router.DetailsData, header: 'TLS', width: '24px' }] : []),
      ...(protocol !== 'udp' ? [{ key: 'rule' as keyof Router.DetailsData, header: 'Rule', sortable: true }] : []),
      { key: 'using', header: 'Entrypoints', sortable: true },
      { key: 'name', header: 'Name', sortable: true },
      { key: 'service', header: 'Service', sortable: true },
      { key: 'provider', header: 'Provider', sortable: true, width: '40px' },
      { key: 'priority', header: 'Priority', sortable: true },
    ]
  }, [protocol])

  if (!routersFound || routersFound.length <= 0) {
    return null
  }

  return (
    <Flex gap={2} css={{ flexDirection: 'column' }}>
      <SectionTitle title="Used by routers" />
      <PaginatedTable
        data={routersFound}
        columns={columns}
        itemsPerPage={10}
        testId="routers-table"
        renderRow={renderRow}
      />
    </Flex>
  )
}
