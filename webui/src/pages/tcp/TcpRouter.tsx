import { Flex, styled, Text } from '@traefiklabs/faency'
import { useParams } from 'react-router-dom'

import { CardListSection, DetailSectionSkeleton } from 'components/resources/DetailSections'
import MiddlewarePanel from 'components/resources/MiddlewarePanel'
import RouterPanel from 'components/resources/RouterPanel'
import TlsPanel from 'components/resources/TlsPanel'
import { ResourceDetailDataType, useResourceDetail } from 'hooks/use-resource-detail'
import Page from 'layout/Page'
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
      <Page title={name}>
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Router right now. Please, try again later.
        </Text>
      </Page>
    )
  }

  if (!data) {
    return (
      <Page title={name}>
        <Flex css={{ flexDirection: 'row', mb: '70px' }} data-testid="skeleton">
          <CardListSection bigDescription />
          <CardListSection />
          <CardListSection isLast />
        </Flex>
        <SpacedColumns>
          <DetailSectionSkeleton />
          <DetailSectionSkeleton />
        </SpacedColumns>
      </Page>
    )
  }

  if (!data.name) {
    return <NotFound />
  }

  return (
    <Page title={name}>
      <RouterStructure data={data} protocol="tcp" />
      <RouterDetail data={data} />
    </Page>
  )
}

export const TcpRouter = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'routers', 'tcp')
  return <TcpRouterRender data={data} error={error} name={name!} />
}

export default TcpRouter
