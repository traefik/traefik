import { Flex, H1, Skeleton, styled, Text } from '@traefiklabs/faency'
import { Helmet } from 'react-helmet-async'

import { ServicePanels } from './ServicePanels'

import { DetailSectionSkeleton } from 'components/resources/DetailSections'
import { UsedByRoutersSection, UsedByRoutersSkeleton } from 'components/resources/UsedByRoutersSection'
import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import { NotFound } from 'pages/NotFound'

type ServiceDetailProps = {
  data?: ResourceDetailDataType
  error?: Error
  name: string
  protocol: 'http' | 'tcp' | 'udp'
}

const SpacedColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))',
  gridGap: '16px',
})

export const ServiceDetail = ({ data, error, name, protocol }: ServiceDetailProps) => {
  if (error) {
    return (
      <>
        <Helmet>
          <title>{name} - Traefik Proxy</title>
        </Helmet>
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Service right now. Please, try again later.
        </Text>
      </>
    )
  }

  if (!data) {
    return (
      <>
        <Helmet>
          <title>{name} - Traefik Proxy</title>
        </Helmet>
        <Skeleton css={{ height: '$7', width: '320px', mb: '$8' }} data-testid="skeleton" />
        <SpacedColumns>
          <DetailSectionSkeleton narrow />
          <DetailSectionSkeleton narrow />
          {protocol !== 'udp' && <DetailSectionSkeleton narrow />}
        </SpacedColumns>
        <UsedByRoutersSkeleton />
      </>
    )
  }

  if (!data.name) {
    return <NotFound />
  }

  return (
    <>
      <Helmet>
        <title>{data.name} - Traefik Proxy</title>
      </Helmet>
      <H1 css={{ mb: '$7' }}>{data.name}</H1>
      <ServicePanels data={data} protocol={protocol} />
      <UsedByRoutersSection data={data} protocol={protocol} />
    </>
  )
}
