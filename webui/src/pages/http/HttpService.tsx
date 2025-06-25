import { Badge, Box, Flex, H1, Skeleton, styled, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { FiGlobe, FiInfo, FiShield } from 'react-icons/fi'
import { useParams } from 'react-router-dom'

import ProviderIcon from 'components/icons/providers'
import {
  BooleanState,
  Chips,
  DetailSection,
  DetailSectionSkeleton,
  ItemBlock,
  ItemTitle,
  LayoutTwoCols,
  ProviderName,
} from 'components/resources/DetailSections'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { UsedByRoutersSection, UsedByRoutersSkeleton } from 'components/resources/UsedByRoutersSection'
import Tooltip from 'components/Tooltip'
import { ResourceDetailDataType, ServiceDetailType, useResourceDetail } from 'hooks/use-resource-detail'
import Page from 'layout/Page'
import { NotFound } from 'pages/NotFound'

type DetailProps = {
  data: ServiceDetailType
  protocol?: string
}

const SpacedColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))',
  gridGap: '16px',
})

const ServicesGrid = styled(Box, {
  display: 'grid',
  gridTemplateColumns: '2fr 1fr 1fr',
  alignItems: 'center',
  padding: '$3 $5',
  borderBottom: '1px solid $tableRowBorder',
})

const ServersGrid = styled(Box, {
  display: 'grid',
  alignItems: 'center',
  padding: '$3 $5',
  borderBottom: '1px solid $tableRowBorder',
})

const MirrorsGrid = styled(Box, {
  display: 'grid',
  gridTemplateColumns: '2fr 1fr 1fr',
  alignItems: 'center',
  padding: '$3 $5',
  borderBottom: '1px solid $tableRowBorder',

  '> *:not(:first-child)': {
    justifySelf: 'flex-end',
  },
})

const GridTitle = styled(Text, {
  fontSize: '14px',
  fontWeight: 700,
  color: 'hsl(0, 0%, 56%)',
})

type Server = {
  url: string
  address?: string
}

type ServerStatus = {
  [server: string]: string
}

function getServerStatusList(data: ServiceDetailType): ServerStatus {
  const serversList: ServerStatus = {}

  data.loadBalancer?.servers?.forEach((server: Server) => {
    serversList[server.address || server.url] = 'DOWN'
  })

  if (data.serverStatus) {
    Object.entries(data.serverStatus).forEach(([server, status]) => {
      serversList[server] = status
    })
  }

  return serversList
}

export const ServicePanels = ({ data, protocol = '' }: DetailProps) => {
  const serversList = getServerStatusList(data)
  const getProviderFromName = (serviceName: string): string => {
    const [, provider] = serviceName.split('@')
    return provider || data.provider
  }
  const providerName = useMemo(() => {
    return data.provider
  }, [data.provider])

  return (
    <SpacedColumns css={{ mb: '$5', pb: '$5' }} data-testid="service-details">
      <DetailSection narrow icon={<FiInfo size={20} />} title="Service Details">
        <LayoutTwoCols>
          {data.type && (
            <ItemBlock title="Type">
              <Text css={{ lineHeight: '32px' }}>{data.type}</Text>
            </ItemBlock>
          )}
          {data.provider && (
            <ItemBlock title="Provider">
              <ProviderIcon name={data.provider} />
              <ProviderName css={{ ml: '$2' }}>{providerName}</ProviderName>
            </ItemBlock>
          )}
        </LayoutTwoCols>
        {data.status && (
          <ItemBlock title="Status">
            <ResourceStatus status={data.status} withLabel />
          </ItemBlock>
        )}
        {data.mirroring && data.mirroring.service && (
          <ItemBlock title="Main Service">
            <Badge>{data.mirroring.service}</Badge>
          </ItemBlock>
        )}
        {data.loadBalancer && (
          <>
            {data.loadBalancer.passHostHeader && (
              <ItemBlock title="Pass Host Header">
                <BooleanState enabled={data.loadBalancer.passHostHeader} />
              </ItemBlock>
            )}
            {data.loadBalancer.terminationDelay && (
              <ItemBlock title="Termination Delay">
                <Text>{`${data.loadBalancer.terminationDelay} ms`}</Text>
              </ItemBlock>
            )}
          </>
        )}
      </DetailSection>
      {data.loadBalancer?.healthCheck && (
        <DetailSection narrow icon={<FiShield size={20} />} title="Health Check">
          <Box data-testid="health-check">
            <LayoutTwoCols>
              {data.loadBalancer.healthCheck.scheme && (
                <ItemBlock title="Scheme">
                  <Text>{data.loadBalancer.healthCheck.scheme}</Text>
                </ItemBlock>
              )}
              {data.loadBalancer.healthCheck.interval && (
                <ItemBlock title="Interval">
                  <Text>{data.loadBalancer.healthCheck.interval}</Text>
                </ItemBlock>
              )}
            </LayoutTwoCols>
            <LayoutTwoCols>
              {data.loadBalancer.healthCheck.path && (
                <ItemBlock title="Path">
                  <Tooltip label={data.loadBalancer.healthCheck.path} action="copy">
                    <Text>{data.loadBalancer.healthCheck.path}</Text>
                  </Tooltip>
                </ItemBlock>
              )}
              {data.loadBalancer.healthCheck.timeout && (
                <ItemBlock title="Timeout">
                  <Text>{data.loadBalancer.healthCheck.timeout}</Text>
                </ItemBlock>
              )}
            </LayoutTwoCols>
            <LayoutTwoCols>
              {data.loadBalancer.healthCheck.port && (
                <ItemBlock title="Port">
                  <Text>{data.loadBalancer.healthCheck.port}</Text>
                </ItemBlock>
              )}
              {data.loadBalancer.healthCheck.hostname && (
                <ItemBlock title="Hostname">
                  <Tooltip label={data.loadBalancer.healthCheck.hostname} action="copy">
                    <Text>{data.loadBalancer.healthCheck.hostname}</Text>
                  </Tooltip>
                </ItemBlock>
              )}
            </LayoutTwoCols>
            {data.loadBalancer.healthCheck.headers && (
              <ItemBlock title="Headers">
                <Chips
                  variant="neon"
                  items={Object.entries(data.loadBalancer.healthCheck.headers).map((entry) => entry.join(': '))}
                />
              </ItemBlock>
            )}
          </Box>
        </DetailSection>
      )}
      {!!data?.weighted?.services?.length && (
        <DetailSection narrow icon={<FiGlobe size={20} />} title="Services" noPadding>
          <>
            <ServicesGrid css={{ mt: '$2' }}>
              <GridTitle>Name</GridTitle>
              <GridTitle css={{ textAlign: 'center' }}>Weight</GridTitle>
              <GridTitle css={{ textAlign: 'center' }}>Provider</GridTitle>
            </ServicesGrid>
            <Box data-testid="servers-list">
              {data.weighted.services.map((service) => (
                <ServicesGrid key={service.name}>
                  <Text>{service.name}</Text>
                  <Text css={{ textAlign: 'center' }}>{service.weight}</Text>
                  <Flex css={{ justifyContent: 'center' }}>
                    <ProviderIcon name={getProviderFromName(service.name)} />
                  </Flex>
                </ServicesGrid>
              ))}
            </Box>
          </>
        </DetailSection>
      )}
      {Object.keys(serversList).length > 0 && (
        <DetailSection narrow icon={<FiGlobe size={20} />} title="Servers" noPadding>
          <>
            <ServersGrid css={{ gridTemplateColumns: protocol === 'http' ? '25% auto' : 'inherit', mt: '$2' }}>
              {protocol === 'http' && <ItemTitle css={{ mb: 0 }}>Status</ItemTitle>}
              <ItemTitle css={{ mb: 0 }}>URL</ItemTitle>
            </ServersGrid>
            <Box data-testid="servers-list">
              {Object.entries(serversList).map(([server, status]) => (
                <ServersGrid key={server} css={{ gridTemplateColumns: protocol === 'http' ? '25% auto' : 'inherit' }}>
                  {protocol === 'http' && <ResourceStatus status={status === 'UP' ? 'enabled' : 'disabled'} />}
                  <Box>
                    <Tooltip label={server} action="copy">
                      <Text>{server}</Text>
                    </Tooltip>
                  </Box>
                </ServersGrid>
              ))}
            </Box>
          </>
        </DetailSection>
      )}
      {data.mirroring?.mirrors && data.mirroring.mirrors.length > 0 && (
        <DetailSection narrow icon={<FiGlobe size={20} />} title="Mirror Services" noPadding>
          <MirrorsGrid css={{ mt: '$2' }}>
            <GridTitle>Name</GridTitle>
            <GridTitle>Percent</GridTitle>
            <GridTitle>Provider</GridTitle>
          </MirrorsGrid>
          <Box data-testid="mirror-services">
            {data.mirroring.mirrors.map((mirror) => (
              <MirrorsGrid key={mirror.name}>
                <Text>{mirror.name}</Text>
                <Text>{mirror.percent}</Text>
                <ProviderIcon name={getProviderFromName(mirror.name)} />
              </MirrorsGrid>
            ))}
          </Box>
        </DetailSection>
      )}
    </SpacedColumns>
  )
}

type HttpServiceRenderProps = {
  data?: ResourceDetailDataType
  error?: Error
  name: string
}

export const HttpServiceRender = ({ data, error, name }: HttpServiceRenderProps) => {
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
      <ServicePanels data={data} protocol="http" />
      <UsedByRoutersSection data={data} protocol="http" />
    </Page>
  )
}

export const HttpService = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'services')
  return <HttpServiceRender data={data} error={error} name={name!} />
}

export default HttpService
