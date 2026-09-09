import { useMemo } from 'react'
import { FiShield } from 'react-icons/fi'

import CopyableText from 'components/CopyableText'
import { Chips } from 'components/resources/DetailItemComponents'
import DetailsCard from 'components/resources/DetailsCard'

type ServiceHealthCheckProps = {
  data: Service.Details
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
            val: <CopyableText text={healthCheck.send} />,
          },
          healthCheck?.expect && {
            key: 'Expect',
            val: <CopyableText text={healthCheck.expect} />,
          },
        ].filter(Boolean) as { key: string; val: string | React.ReactElement; stackVertical?: boolean }[]
      } else {
        return [
          healthCheck?.scheme && { key: 'Scheme', val: healthCheck.scheme },
          healthCheck?.interval && { key: 'Interval', val: healthCheck.interval },
          healthCheck?.path && {
            key: 'Path',
            val: <CopyableText text={data.loadBalancer.healthCheck.path} />,
          },
          healthCheck?.timeout && { key: 'Timeout', val: healthCheck.timeout },
          healthCheck?.port && { key: 'Port', val: String(healthCheck.port) },
          healthCheck?.hostname && {
            key: 'Hostname',
            val: <CopyableText text={data.loadBalancer.healthCheck.hostname} />,
          },
          healthCheck.headers && {
            key: 'Headers',
            val: <Chips variant="neon" items={Object.entries(healthCheck.headers).map((entry) => entry.join(': '))} />,
            stackVertical: true,
            forceNewRow: true,
          },
        ].filter(Boolean) as { key: string; val: string | React.ReactElement; stackVertical?: boolean }[]
      }
    }
  }, [data.loadBalancer?.healthCheck, isTcp])

  if (!healthCheckItems) return null

  return (
    <DetailsCard icon={<FiShield size={20} />} title="Health Check" items={healthCheckItems} testId="health-check" />
  )
}

export default ServiceHealthCheck
