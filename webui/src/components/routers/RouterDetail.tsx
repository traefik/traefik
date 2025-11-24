import { Flex, styled, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { Helmet } from 'react-helmet-async'

import { CardListSection, DetailSectionSkeleton } from 'components/resources/DetailSections'
import { RouterPanels } from 'components/routers/RouterPanels'
import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import { NotFound } from 'pages/NotFound'

type RouterDetailProps = {
  data?: ResourceDetailDataType
  error?: Error | null
  name: string
  protocol: 'http' | 'tcp' | 'udp'
}

const SpacedColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))',
  gridGap: '16px',
})

export const RouterDetail = ({ data, error, name, protocol }: RouterDetailProps) => {
  const isHttp = useMemo(() => protocol === 'http', [protocol])

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
    const skeletonCount = isHttp ? 3 : 2

    return (
      <>
        <Helmet>
          <title>{name} - Traefik Proxy</title>
        </Helmet>
        <Flex css={{ flexDirection: 'row', mb: '70px' }} data-testid="skeleton">
          <CardListSection bigDescription />
          <CardListSection />
          {isHttp && <CardListSection />}
          <CardListSection isLast />
        </Flex>
        <SpacedColumns>
          {Array.from({ length: skeletonCount }).map((_, i) => (
            <DetailSectionSkeleton key={i} />
          ))}
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
      <RouterPanels data={data} protocol={protocol} />
    </>
  )
}
