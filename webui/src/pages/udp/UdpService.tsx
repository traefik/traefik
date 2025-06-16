import { Flex, H1, Skeleton, styled, Text } from '@traefiklabs/faency'
import { useParams } from 'react-router-dom'

import { DetailSectionSkeleton } from 'components/resources/DetailSections'
import { UsedByRoutersSection, UsedByRoutersSkeleton } from 'components/resources/UsedByRoutersSection'
import { ResourceDetailDataType, useResourceDetail } from 'hooks/use-resource-detail'
import Page from 'layout/Page'
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
      <Page title={name}>
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Service right now. Please, try again later.
        </Text>
      </Page>
    )
  }

  if (!data) {
    return (
      <Page title={name}>
        <Skeleton css={{ height: '$7', width: '320px', mb: '$8' }} data-testid="skeleton" />
        <SpacedColumns>
          <DetailSectionSkeleton narrow />
          <DetailSectionSkeleton narrow />
        </SpacedColumns>
        <UsedByRoutersSkeleton />
      </Page>
    )
  }

  if (!data.name) {
    return <NotFound />
  }

  return (
    <Page title={name}>
      <H1 css={{ mb: '$7' }}>{data.name}</H1>
      <ServicePanels data={data} />
      <UsedByRoutersSection data={data} protocol="udp" />
    </Page>
  )
}

export const UdpService = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'services', 'udp')
  return <UdpServiceRender data={data} error={error} name={name!} />
}

export default UdpService
