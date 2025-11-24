import { Badge, Box, Flex, styled, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { FiGlobe, FiInfo, FiShield } from 'react-icons/fi'

import ProviderIcon from 'components/icons/providers'
import {
  BooleanState,
  Chips,
  DetailSection,
  ItemBlock,
  ItemTitle,
  LayoutTwoCols,
  ProviderName,
} from 'components/resources/DetailSections'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import Tooltip from 'components/Tooltip'
import { ServiceDetailType } from 'hooks/use-resource-detail'

type ServicePanelsProps = {
  data: ServiceDetailType
  protocol: 'http' | 'tcp' | 'udp'
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
  url?: string
  address?: string
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

function getServerStatusList(data: ServiceDetailType): ServerStatus {
  const serversList: ServerStatus = {}

  data.loadBalancer?.servers?.forEach((server: Server) => {
    const serverKey = server.address || server.url
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

export const ServicePanels = ({ data, protocol }: ServicePanelsProps) => {
  const serversList = getServerStatusList(data)
  const getProviderFromName = (serviceName: string): string => {
    const [, provider] = serviceName.split('@')
    return provider || data.provider
  }
  const providerName = useMemo(() => {
    return data.provider
  }, [data.provider])

  const isTcp = useMemo(() => protocol === 'tcp', [protocol])
  const isUdp = useMemo(() => protocol === 'udp', [protocol])

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
            {isTcp ? (
              // TCP Health Check
              (() => {
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
              })()
            ) : (
              // HTTP/UDP Health Check
              <>
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
              </>
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
            <Box data-testid="weighted-services">
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
            <ServersGrid css={{ gridTemplateColumns: !isUdp ? '25% auto' : 'inherit', mt: '$2' }}>
              {!isUdp && <ItemTitle css={{ mb: 0 }}>Status</ItemTitle>}
              <ItemTitle css={{ mb: 0 }}>{isTcp ? 'Address' : 'URL'}</ItemTitle>
            </ServersGrid>
            <Box data-testid="servers-list">
              {Object.entries(serversList).map(([server, status]) => (
                <ServersGrid key={server} css={{ gridTemplateColumns: !isUdp ? '25% auto' : 'inherit' }}>
                  {!isUdp && <ResourceStatus status={status === 'UP' ? 'enabled' : 'disabled'} />}
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
