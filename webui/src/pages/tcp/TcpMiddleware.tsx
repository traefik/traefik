import { Card, Box, H1, Skeleton, styled, Text } from '@traefiklabs/faency'
import { Helmet } from 'react-helmet-async'
import { useParams } from 'react-router-dom'

import { DetailSectionSkeleton } from 'components/resources/DetailSections'
import { RenderMiddleware } from 'components/resources/MiddlewarePanel'
import { UsedByRoutersSection, UsedByRoutersSkeleton } from 'components/resources/UsedByRoutersSection'
import { ResourceDetailDataType, useResourceDetail } from 'hooks/use-resource-detail'
import { NotFound } from 'pages/NotFound'
import breakpoints from 'utils/breakpoints'

const MiddlewareGrid = styled(Box, {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(400px, 1fr))',

  [`@media (max-width: ${breakpoints.tablet})`]: {
    gridTemplateColumns: '1fr',
  },
})

type TcpMiddlewareRenderProps = {
  data?: ResourceDetailDataType
  error?: Error
  name: string
}

export const TcpMiddlewareRender = ({ data, error, name }: TcpMiddlewareRenderProps) => {
  if (error) {
    return (
      <>
        <Helmet>
          <title>{name} - Traefik Proxy</title>
        </Helmet>
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Middleware right now. Please, try again later.
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
        <Skeleton css={{ height: '$7', width: '320px', mb: '$4' }} data-testid="skeleton" />
        <MiddlewareGrid>
          <DetailSectionSkeleton />
        </MiddlewareGrid>
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
      <MiddlewareGrid>
        <Card css={{ padding: '$5' }} data-testid="middleware-card">
          <RenderMiddleware middleware={data} />
        </Card>
      </MiddlewareGrid>
      <UsedByRoutersSection data-testid="routers-table" data={data} protocol="tcp" />
    </>
  )
}

export const TcpMiddleware = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'middlewares', 'tcp')
  return <TcpMiddlewareRender data={data} error={error} name={name!} />
}

export default TcpMiddleware
