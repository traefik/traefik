import { Text } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { FiShield } from 'react-icons/fi'

import DetailsCard from 'components/resources/DetailsCard'
import { Chips } from 'components/resources/DetailSections'
import Tooltip from 'components/Tooltip'
import { ServiceDetailType } from 'hooks/use-resource-detail'

type ServiceHealthCheckProps = {
  data: ServiceDetailType
  protocol: 'http' | 'tcp' | 'udp'
}

const ServiceHealthCheck = ({ data, protocol }: ServiceHealthCheckProps) => {
  const isTcp = useMemo(() => protocol === 'tcp', [protocol])

  const healthCheckItems = useMemo(() => {
    if (data.loadBalancer?.healthCheck) {
      const healthCheck = data.loadBalancer.healthCheck
      if (isTcp) {
        return [
          healthCheck?.interval && { key: 'Interval', val: healthCheck.interval },
          healthCheck?.timeout && { key: 'Timeout', val: healthCheck.timeout },
          healthCheck?.port && { key: 'Port', val: healthCheck.port },
          healthCheck?.unhealthyInterval && { key: 'Unhealthy interval', val: healthCheck.unhealthyInterval },
          healthCheck?.send && {
            key: 'Send',
            val: (
              <Tooltip label={healthCheck.send} action="copy">
                <Text>{healthCheck.send}</Text>
              </Tooltip>
            ),
          },
          healthCheck?.expect && {
            key: 'Expect',
            val: (
              <Tooltip label={healthCheck.expect} action="copy">
                <Text>{healthCheck.expect}</Text>
              </Tooltip>
            ),
          },
        ].filter(Boolean) as { key: string; val: string | React.ReactElement; stackVertical?: boolean }[]
      } else {
        return [
          healthCheck?.scheme && { key: 'Scheme', val: healthCheck.scheme },
          healthCheck?.interval && { key: 'Interval', val: healthCheck.interval },
          healthCheck?.path && {
            key: 'Path',
            val: (
              <Tooltip label={data.loadBalancer.healthCheck.path} action="copy">
                <Text>{data.loadBalancer.healthCheck.path}</Text>
              </Tooltip>
            ),
          },
          healthCheck?.timeout && { key: 'Timeout', val: healthCheck.timeout },
          healthCheck?.port && { key: 'Port', val: healthCheck.port },
          healthCheck?.hostname && {
            key: 'Hostname',
            val: (
              <Tooltip label={data.loadBalancer.healthCheck.hostname} action="copy">
                <Text>{data.loadBalancer.healthCheck.hostname}</Text>
              </Tooltip>
            ),
          },
          healthCheck.headers && {
            key: 'Headers',
            val: <Chips variant="neon" items={Object.entries(healthCheck.headers).map((entry) => entry.join(': '))} />,
            stackVertical: true,
          },
        ].filter(Boolean) as { key: string; val: string | React.ReactElement; stackVertical?: boolean }[]
      }
    }
  }, [data.loadBalancer?.healthCheck, isTcp])

  if (!healthCheckItems) return null

  return <DetailsCard icon={<FiShield size={20} />} title="Health Check" items={healthCheckItems} />
}

export default ServiceHealthCheck
