import { Flex, styled, Text } from '@traefiklabs/faency'
import { useParams } from 'react-router-dom'

import { CardListSection, DetailSectionSkeleton } from 'components/resources/DetailSections'
import RouterPanel from 'components/resources/RouterPanel'
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
      <RouterStructure data={data} protocol="udp" />
      <RouterDetail data={data} />
    </Page>
  )
}

export const UdpRouter = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'routers', 'udp')

  return <UdpRouterRender data={data} error={error} name={name!} />
}

export default UdpRouter
