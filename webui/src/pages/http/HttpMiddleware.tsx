import { Box, H1, Skeleton, styled, Text } from '@traefiklabs/faency'
import { Card } from 'components/FaencyOverrides'
import { DetailSectionSkeleton } from 'components/resources/DetailSections'
import { RenderMiddleware } from 'components/resources/MiddlewarePanel'
import { UsedByRoutersSection, UsedByRoutersSkeleton } from 'components/resources/UsedByRoutersSection'
import { ResourceDetailDataType, useResourceDetail } from 'hooks/use-resource-detail'
import Page from 'layout/Page'
import { NotFound } from 'pages/NotFound'
import { useParams } from 'react-router-dom'
import breakpoints from 'utils/breakpoints'

const MiddlewareGrid = styled(Box, {
  display: 'grid',
  gridTemplateColumns: 'repeat(3, minmax(400px, 1fr))',

  [`@media (max-width: ${breakpoints.tablet})`]: {
    gridTemplateColumns: '1fr',
  },
})

type HttpMiddlewareRenderProps = {
  data?: ResourceDetailDataType
  error?: any // eslint-disable-line @typescript-eslint/no-explicit-any
  name: string
}

export const HttpMiddlewareRender = ({ data, error, name }: HttpMiddlewareRenderProps) => {
  if (error) {
    return (
      <Page title={name}>
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Middleware right now. Please, try again later.
        </Text>
      </Page>
    )
  }

  if (!data) {
    return (
      <Page title={name}>
        <Skeleton css={{ height: '$7', width: '320px', mb: '$4' }} data-testid="skeleton" />
        <MiddlewareGrid data-testid="skeletons">
          <DetailSectionSkeleton />
        </MiddlewareGrid>
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
      <MiddlewareGrid>
        <Card css={{ p: '$3' }} data-testid="middleware-card">
          <RenderMiddleware middleware={data} />
        </Card>
      </MiddlewareGrid>
      <UsedByRoutersSection data-testid="routers-table" data={data} protocol="http" />
    </Page>
  )
}

export const HttpMiddleware = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'middlewares')
  return <HttpMiddlewareRender data={data} error={error} name={name!} />
}

export default HttpMiddleware
