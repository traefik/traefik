import { Box, Flex, H1, Skeleton, Text } from '@traefiklabs/faency'

import MirrorServices from './MirrorServices'
import Servers from './Servers'
import ServiceDefinition from './ServiceDefinition'
import ServiceHealthCheck from './ServiceHealthCheck'
import WeightedServices from './WeightedServices'

import { DetailsCardSkeleton } from 'components/resources/DetailsCard'
import { UsedByRoutersSection, UsedByRoutersSkeleton } from 'components/resources/UsedByRoutersSection'
import AriaTableSkeleton from 'components/tables/AriaTableSkeleton'
import PageTitle from 'layout/PageTitle'
import { NotFound } from 'pages/NotFound'

type ServiceDetailProps = {
  data?: Resource.DetailsData
  error?: Error
  name: string
  protocol: 'http' | 'tcp' | 'udp'
}

export const ServiceDetail = ({ data, error, name, protocol }: ServiceDetailProps) => {
  if (error) {
    return (
      <>
        <PageTitle title={data?.name || name} />
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Service right now. Please, try again later.
        </Text>
      </>
    )
  }

  if (!data) {
    return (
      <>
        <PageTitle title={name} />
        <Skeleton css={{ height: '$7', width: '320px', mb: '$7' }} data-testid="skeleton" />
        <Flex direction="column" gap={4}>
          <DetailsCardSkeleton />
          <DetailsCardSkeleton />

          <Box>
            <Skeleton css={{ height: '$5', width: '150px', mb: '$2' }} />
            <AriaTableSkeleton columns={2} />
          </Box>
          <Box>
            <Skeleton css={{ height: '$5', width: '150px', mb: '$2' }} />
            <AriaTableSkeleton />
          </Box>

          <UsedByRoutersSkeleton />
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
        <ServiceDefinition data={data} testId="service-details" />

        {data.loadBalancer?.healthCheck && <ServiceHealthCheck data={data} protocol={protocol} />}
        {!!data?.weighted?.services?.length && (
          <WeightedServices services={data.weighted.services} defaultProvider={data.provider} />
        )}
        <Servers data={data} protocol={protocol} />
        {!!data?.mirroring?.mirrors && (
          <MirrorServices mirrors={data.mirroring?.mirrors} defaultProvider={data.provider} />
        )}
        <UsedByRoutersSection data={data} protocol={protocol} />
      </Flex>
    </>
  )
}
