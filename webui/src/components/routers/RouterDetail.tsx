import { Flex, H1, Skeleton, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'

import { DetailsCardSkeleton } from 'components/resources/DetailsCard'
import ResourceErrors, { ResourceErrorsSkeleton } from 'components/resources/ResourceErrors'
import RouterFlowDiagram, { RouterFlowDiagramSkeleton } from 'components/routers/RouterFlowDiagram'
import TlsSection from 'components/routers/TlsSection'
import PageTitle from 'layout/PageTitle'
import { NotFound } from 'pages/NotFound'

type RouterDetailProps = {
  data?: Resource.DetailsData
  error?: Error | null
  name: string
  protocol: 'http' | 'tcp' | 'udp'
}

export const RouterDetail = ({ data, error, name, protocol }: RouterDetailProps) => {
  const isUdp = useMemo(() => protocol === 'udp', [protocol])

  if (error) {
    return (
      <>
        <PageTitle title={data?.name || name} />
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Router right now. Please, try again later.
        </Text>
      </>
    )
  }

  if (!data) {
    return (
      <>
        <PageTitle title={name} />
        <Skeleton css={{ height: '$7', width: '320px', mb: '$7' }} data-testid="skeleton" />
        <Flex direction="column" gap={6}>
          <RouterFlowDiagramSkeleton />
          <ResourceErrorsSkeleton />
          <DetailsCardSkeleton />
        </Flex>
      </>
    )
  }

  if (!data.name) {
    return <NotFound />
  }

  return (
    <>
      <PageTitle title={data.name} />
      <H1 css={{ mb: '$7' }}>{data.name}</H1>
      <Flex direction="column" gap={6}>
        <RouterFlowDiagram data={data} protocol={protocol} />
        {data?.error && <ResourceErrors errors={data.error} />}
        {!isUdp && <TlsSection data={data?.tls} />}
      </Flex>
    </>
  )
}
