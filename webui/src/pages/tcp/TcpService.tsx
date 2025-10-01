import { Box, Flex, H1, Skeleton, styled, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { Helmet } from 'react-helmet-async'
import { FiGlobe, FiInfo, FiShield } from 'react-icons/fi'
import { useParams } from 'react-router-dom'

import ProviderIcon from 'components/icons/providers'
import {
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
import { NotFound } from 'pages/NotFound'

type TcpDetailProps = {
  data: ServiceDetailType
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

const GridTitle = styled(Text, {
  fontSize: '14px',
  fontWeight: 700,
  color: 'hsl(0, 0%, 56%)',
})

type TcpServer = {
  address: string
}

type ServerStatus = {
  [server: string]: string
}

type TcpHealthCheck = {
  port?: number
  send?: string
  expect?: string
  interval?: string
  unhealthyInterval?: string
  timeout?: string
}

function getTcpServerStatusList(data: ServiceDetailType): ServerStatus {
  const serversList: ServerStatus = {}

  data.loadBalancer?.servers?.forEach((server: any) => {
    // TCP servers should have address, but handle both url and address for compatibility
    const serverKey = (server as TcpServer).address || (server as any).url
    if (serverKey) {
      serversList[serverKey] = 'DOWN'
    }
  })

  if (data.serverStatus) {
    Object.entries(data.serverStatus).forEach(([server, status]) => {
      serversList[server] = status
    })
  }

  return serversList
}

export const TcpServicePanels = ({ data }: TcpDetailProps) => {
  const serversList = getTcpServerStatusList(data)
  const getProviderFromName = (serviceName: string): string => {
    const [, provider] = serviceName.split('@')
    return provider || data.provider
  }
  const providerName = useMemo(() => {
    return data.provider
  }, [data.provider])

  return (
    <SpacedColumns css={{ mb: '$5', pb: '$5' }} data-testid="tcp-service-details">
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
        {data.loadBalancer && (
          <>
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
          <Box data-testid="tcp-health-check">
            {(() => {
              const tcpHealthCheck = data.loadBalancer.healthCheck as unknown as TcpHealthCheck
              return (
                <>
                  <LayoutTwoCols>
                    {tcpHealthCheck.interval && (
                      <ItemBlock title="Interval">
                        <Text>{tcpHealthCheck.interval}</Text>
                      </ItemBlock>
                    )}
                    {tcpHealthCheck.timeout && (
                      <ItemBlock title="Timeout">
                        <Text>{tcpHealthCheck.timeout}</Text>
                      </ItemBlock>
                    )}
                  </LayoutTwoCols>
                  <LayoutTwoCols>
                    {tcpHealthCheck.port && (
                      <ItemBlock title="Port">
                        <Text>{tcpHealthCheck.port}</Text>
                      </ItemBlock>
                    )}
                    {tcpHealthCheck.unhealthyInterval && (
                      <ItemBlock title="Unhealthy Interval">
                        <Text>{tcpHealthCheck.unhealthyInterval}</Text>
                      </ItemBlock>
                    )}
                  </LayoutTwoCols>
                  <LayoutTwoCols>
                    {tcpHealthCheck.send && (
                      <ItemBlock title="Send">
                        <Tooltip label={tcpHealthCheck.send} action="copy">
                          <Text>{tcpHealthCheck.send}</Text>
                        </Tooltip>
                      </ItemBlock>
                    )}
                    {tcpHealthCheck.expect && (
                      <ItemBlock title="Expect">
                        <Tooltip label={tcpHealthCheck.expect} action="copy">
                          <Text>{tcpHealthCheck.expect}</Text>
                        </Tooltip>
                      </ItemBlock>
                    )}
                  </LayoutTwoCols>
                </>
              )
            })()}
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
            <Box data-testid="tcp-weighted-services">
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
            <ServersGrid css={{ gridTemplateColumns: '25% auto', mt: '$2' }}>
              <ItemTitle css={{ mb: 0 }}>Status</ItemTitle>
              <ItemTitle css={{ mb: 0 }}>Address</ItemTitle>
            </ServersGrid>
            <Box data-testid="tcp-servers-list">
              {Object.entries(serversList).map(([server, status]) => (
                <ServersGrid key={server} css={{ gridTemplateColumns: '25% auto' }}>
                  <ResourceStatus status={status === 'UP' ? 'enabled' : 'disabled'} />
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
    </SpacedColumns>
  )
}

type TcpServiceRenderProps = {
  data?: ResourceDetailDataType
  error?: Error
  name: string
}

export const TcpServiceRender = ({ data, error, name }: TcpServiceRenderProps) => {
  if (error) {
    return (
      <>
        <Helmet>
          <title>{name} - Traefik Proxy</title>
        </Helmet>
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Service right now. Please, try again later.
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
        <Skeleton css={{ height: '$7', width: '320px', mb: '$8' }} data-testid="skeleton" />
        <SpacedColumns>
          <DetailSectionSkeleton narrow />
          <DetailSectionSkeleton narrow />
          <DetailSectionSkeleton narrow />
        </SpacedColumns>
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
      <TcpServicePanels data={data} />
      <UsedByRoutersSection data={data} protocol="tcp" />
    </>
  )
}

export const TcpService = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'services', 'tcp')
  return <TcpServiceRender data={data} error={error} name={name!} />
}

export default TcpService
