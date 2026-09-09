import { useMemo } from 'react'

import ProviderIcon from 'components/icons/providers'
import { ProviderName } from 'components/resources/DetailItemComponents'
import DetailsCard from 'components/resources/DetailsCard'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import { parseMiddlewareType } from 'libs/parsers'

type MiddlewareDefinitionProps = {
  data: Middleware.Details
  testId?: string
}

const MiddlewareDefinition = ({ data, testId }: MiddlewareDefinitionProps) => {
  const providerName = useMemo(() => {
    return data.provider
  }, [data.provider])

  const detailsItems = useMemo(
    () =>
      [
        data.status && { key: 'Status', val: <ResourceStatus status={data.status} withLabel /> },
        (data.type || data.plugin) && { key: 'Type', val: parseMiddlewareType(data) },
        data.provider && {
          key: 'Provider',
          val: (
            <>
              <ProviderIcon name={data.provider} />
              <ProviderName css={{ ml: '$2' }}>{providerName}</ProviderName>
            </>
          ),
        },
      ].filter(Boolean) as { key: string; val: string | React.ReactElement }[],
    [data, providerName],
  )

  return <DetailsCard items={detailsItems} testId={testId} />
}

export default MiddlewareDefinition
