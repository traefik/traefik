import { Flex, styled, Text } from '@traefiklabs/faency'
import { Helmet } from 'react-helmet-async'
import { useParams } from 'react-router-dom'

import { CardListSection, DetailSectionSkeleton } from 'components/resources/DetailSections'
import MiddlewarePanel from 'components/resources/MiddlewarePanel'
import RouterPanel from 'components/resources/RouterPanel'
import TlsPanel from 'components/resources/TlsPanel'
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
    <TlsPanel data={data} />
    <MiddlewarePanel data={data} />
  </SpacedColumns>
)

type TcpRouterRenderProps = {
  data?: ResourceDetailDataType
  error?: Error
  name: string
}

export const TcpRouterRender = ({ data, error, name }: TcpRouterRenderProps) => {
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
      <RouterStructure data={data} protocol="tcp" />
      <RouterDetail data={data} />
    </>
  )
}

export const TcpRouter = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'routers', 'tcp')
  return <TcpRouterRender data={data} error={error} name={name!} />
}

export default TcpRouter
