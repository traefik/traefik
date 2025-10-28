import { Flex, H1, Skeleton, styled, Text } from '@traefiklabs/faency'
import { Helmet } from 'react-helmet-async'
import { useParams } from 'react-router-dom'

import { DetailSectionSkeleton } from 'components/resources/DetailSections'
import { UsedByRoutersSection, UsedByRoutersSkeleton } from 'components/resources/UsedByRoutersSection'
import { ResourceDetailDataType, useResourceDetail } from 'hooks/use-resource-detail'
import { ServicePanels } from 'pages/http/HttpService'
import { NotFound } from 'pages/NotFound'

const SpacedColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))',
  gridGap: '16px',
})

type UdpServiceRenderProps = {
  data?: ResourceDetailDataType
  error?: Error
  name: string
}

export const UdpServiceRender = ({ data, error, name }: UdpServiceRenderProps) => {
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
      <ServicePanels data={data} />
      <UsedByRoutersSection data={data} protocol="udp" />
    </>
  )
}

export const UdpService = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'services', 'udp')
  return <UdpServiceRender data={data} error={error} name={name!} />
}

export default UdpService
