import { Badge } from '@traefiklabs/faency'
import { useMemo } from 'react'

import ProviderIcon from 'components/icons/providers'
import { BooleanState, ProviderName } from 'components/resources/DetailItemComponents'
import DetailsCard from 'components/resources/DetailsCard'
import { ResourceStatus } from 'components/resources/ResourceStatus'

type ServiceDefinitionProps = {
  data: Service.Details
  testId?: string
}

const ServiceDefinition = ({ data, testId }: ServiceDefinitionProps) => {
  const providerName = useMemo(() => {
    return data.provider
  }, [data.provider])

  const detailsItems = useMemo(
    () =>
      [
        data.status && { key: 'Status', val: <ResourceStatus status={data.status} withLabel /> },
        data.type && { key: 'Type', val: data.type },
        data.provider && {
          key: 'Provider',
          val: (
            <>
              <ProviderIcon name={data.provider} />
              <ProviderName css={{ ml: '$2' }}>{providerName}</ProviderName>
            </>
          ),
        },
        data.mirroring &&
          data.mirroring.service && { key: 'Main service', val: <Badge>{data.mirroring.service}</Badge> },
        data.loadBalancer?.passHostHeader && {
          key: 'Pass host header',
          val: <BooleanState enabled={data.loadBalancer.passHostHeader} />,
        },
        data.loadBalancer?.terminationDelay && {
          key: 'Termination delay',
          val: `${data.loadBalancer.terminationDelay} ms`,
        },
      ].filter(Boolean) as { key: string; val: string | React.ReactElement }[],
    [
      data.loadBalancer?.passHostHeader,
      data.loadBalancer?.terminationDelay,
      data.mirroring,
      data.provider,
      data.status,
      data.type,
      providerName,
    ],
  )

  return <DetailsCard items={detailsItems} testId={testId} />
}

export default ServiceDefinition
