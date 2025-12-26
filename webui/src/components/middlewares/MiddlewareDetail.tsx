import { Card, Flex, H1, Skeleton, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'

import MiddlewareDefinition from './MiddlewareDefinition'
import { RenderUnknownProp } from './RenderUnknownProp'

import { DetailsCardSkeleton } from 'components/resources/DetailsCard'
import ResourceErrors, { ResourceErrorsSkeleton } from 'components/resources/ResourceErrors'
import { UsedByRoutersSection, UsedByRoutersSkeleton } from 'components/resources/UsedByRoutersSection'
import PageTitle from 'layout/PageTitle'
import { NotFound } from 'pages/NotFound'

type MiddlewareDetailProps = {
  data?: Resource.DetailsData
  error?: Error | null
  name: string
  protocol: 'http' | 'tcp'
}

const filterMiddlewareProps = (middleware: Middleware.Details): string[] => {
  const filteredProps = [] as string[]
  const propsToRemove = ['name', 'plugin', 'status', 'type', 'provider', 'error', 'usedBy', 'routers']

  Object.keys(middleware).map((propName) => {
    if (!propsToRemove.includes(propName)) {
      filteredProps.push(propName)
    }
  })

  return filteredProps
}

export const MiddlewareDetail = ({ data, error, name, protocol }: MiddlewareDetailProps) => {
  const filteredProps = useMemo(() => {
    if (data) {
      return filterMiddlewareProps(data)
    }

    return []
  }, [data])

  if (error) {
    return (
      <>
        <PageTitle title={data?.name || name} />
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Middleware right now. Please, try again later.
        </Text>
      </>
    )
  }

  if (!data) {
    return (
      <>
        <PageTitle title={name} />
        <Skeleton css={{ height: '$7', width: '320px', mb: '$7' }} data-testid="skeleton" />
        <Flex direction="column" gap={6}>
          <DetailsCardSkeleton />
          <ResourceErrorsSkeleton />
          <UsedByRoutersSkeleton />
        </Flex>
      </>
    )
  }

  if (!data.name) {
    return <NotFound />
  }

  return (
    <>
      <PageTitle title={data.name} />
      <H1 css={{ mb: '$7' }}>{data.name}</H1>
      <Flex direction="column" gap={6}>
        <MiddlewareDefinition data={data} testId="middleware-card" />
        {!!data.error && <ResourceErrors errors={data.error} />}
        {(!!data.plugin || !!filteredProps.length) && (
          <Card>
            {data.plugin &&
              Object.keys(data.plugin).map((pluginName) => (
                <RenderUnknownProp key={pluginName} name={pluginName} prop={data.plugin?.[pluginName]} />
              ))}
            {filteredProps?.map((propName) => (
              <RenderUnknownProp key={propName} name={propName} prop={data[propName]} removeTitlePrefix={data.type} />
            ))}
          </Card>
        )}

        <UsedByRoutersSection data-testid="routers-table" data={data} protocol={protocol} />
      </Flex>
    </>
  )
}
