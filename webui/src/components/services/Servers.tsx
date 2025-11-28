import { Flex, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { FiGlobe } from 'react-icons/fi'

import { SectionTitle } from 'components/resources/DetailsCard'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { colorByStatus } from 'components/resources/Status'
import PaginatedTable from 'components/tables/PaginatedTable'
import Tooltip from 'components/Tooltip'

type ServersProps = {
  data: Service.Details
  protocol: 'http' | 'tcp' | 'udp'
}

type Server = {
  url?: string
  address?: string
}

type ServerStatus = {
  [server: string]: string
}

function getServerStatusList(data: Service.Details): ServerStatus {
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

export const getProviderFromName = (serviceName: string, defaultProvider: string): string => {
  const [, provider] = serviceName.split('@')
  return provider || defaultProvider
}

const Servers = ({ data, protocol }: ServersProps) => {
  const serversList = getServerStatusList(data)

  const isTcp = useMemo(() => protocol === 'tcp', [protocol])
  const isUdp = useMemo(() => protocol === 'udp', [protocol])

  if (!Object.keys(serversList)?.length) return null

  return (
    <Flex direction="column" gap={2}>
      <SectionTitle icon={<FiGlobe size={20} />} title="Servers" />
      <PaginatedTable
        data={Object.entries(serversList).map(([server, status]) => ({
          server,
          status,
        }))}
        columns={[
          ...(isUdp ? [] : [{ key: 'status' as const, header: 'Status' }]),
          { key: 'server' as const, header: isTcp ? 'Address' : 'URL' },
        ]}
        testId="servers-list"
        renderCell={(key, value) => {
          if (key === 'status') {
            return (
              <Flex align="center" gap={2}>
                <ResourceStatus status={value === 'UP' ? 'enabled' : 'disabled'} />
                <Text css={{ color: value === 'UP' ? colorByStatus.success : colorByStatus.disabled }}>{value}</Text>
              </Flex>
            )
          }
          if (key === 'server') {
            return (
              <Tooltip label={value} action="copy">
                <Text>{value}</Text>
              </Tooltip>
            )
          }
          return <Text>{value}</Text>
        }}
      />
    </Flex>
  )
}

export default Servers
