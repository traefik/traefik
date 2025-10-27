import { Flex, styled, Text } from '@traefiklabs/faency'
import { Helmet } from 'react-helmet-async'
import { useParams } from 'react-router-dom'

import { CardListSection, DetailSectionSkeleton } from 'components/resources/DetailSections'
import RouterPanel from 'components/resources/RouterPanel'
import { ResourceDetailDataType, useResourceDetail } from 'hooks/use-resource-detail'
import { RouterStructure } from 'pages/http/HttpRouter'
import { NotFound } from 'pages/NotFound'

type DetailProps = {
  data: ResourceDetailDataType
}

const SpacedColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))',
  gridGap: '16px',
})

const RouterDetail = ({ data }: DetailProps) => (
  <SpacedColumns data-testid="router-details">
    <RouterPanel data={data} />
  </SpacedColumns>
)

type UdpRouterRenderProps = {
  data?: ResourceDetailDataType
  error?: Error
  name: string
}

export const UdpRouterRender = ({ data, error, name }: UdpRouterRenderProps) => {
  if (error) {
    return (
      <>
        <Helmet>
          <title>{name} - Traefik Proxy</title>
        </Helmet>
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Router right now. Please, try again later.
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
        <Flex css={{ flexDirection: 'row', mb: '70px' }} data-testid="skeleton">
          <CardListSection bigDescription />
          <CardListSection />
          <CardListSection isLast />
        </Flex>
        <SpacedColumns>
          <DetailSectionSkeleton />
          <DetailSectionSkeleton />
        </SpacedColumns>
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
      <RouterStructure data={data} protocol="udp" />
      <RouterDetail data={data} />
    </>
  )
}

export const UdpRouter = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'routers', 'udp')

  return <UdpRouterRender data={data} error={error} name={name!} />
}

export default UdpRouter
