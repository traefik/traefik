import { Flex, Text } from '@traefik-labs/faency'
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
  weight?: number
}

function getServerStatusList(data: Service.Details) {
  if (!data?.loadBalancer?.servers) {
    return []
  }
  return data.loadBalancer?.servers?.map((server: Server) => ({
    url: server.address || server.url,
    status: data.serverStatus?.[server.address || server.url || '-'] || 'DOWN',
    weight: server.weight ?? 1,
  }))
}

export const getProviderFromName = (serviceName: string, defaultProvider: string): string => {
  const [, provider] = serviceName.split('@')
  return provider || defaultProvider
}

const Servers = ({ data, protocol }: ServersProps) => {
  const serversList = useMemo(() => getServerStatusList(data), [data])

  const isTcp = useMemo(() => protocol === 'tcp', [protocol])
  const isUdp = useMemo(() => protocol === 'udp', [protocol])

  if (!Object.keys(serversList)?.length) return null

  return (
    <Flex direction="column" gap={2}>
      <SectionTitle icon={<FiGlobe size={20} />} title="Servers" />
      {serversList?.length > 0 && (
        <PaginatedTable
          data={serversList?.map(({ url, status, weight }) => ({
            server: url,
            status,
            weight,
          }))}
          columns={[
            ...(isUdp ? [] : [{ key: 'status' as const, header: 'Status' }]),
            { key: 'server' as const, header: isTcp ? 'Address' : 'URL' },
            ...(isUdp ? [] : [{ key: 'weight' as const, header: 'Weight' }]),
          ]}
          testId={`${protocol}-servers-list`}
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
                <Tooltip label={value as string} action="copy">
                  <Text>{value}</Text>
                </Tooltip>
              )
            }
            return <Text>{value}</Text>
          }}
        />
      )}
    </Flex>
  )
}

export default Servers
