import { Badge, Text } from '@traefiklabs/faency'
import { FiInfo } from 'react-icons/fi'

import { DetailSection, ItemBlock, LayoutTwoCols, ProviderName } from './DetailSections'
import GenericTable from './GenericTable'
import { ResourceStatus } from './ResourceStatus'

import ProviderIcon from 'components/icons/providers'
import Tooltip from 'components/Tooltip'
import { ResourceDetailDataType } from 'hooks/use-resource-detail'

type Props = {
  data: ResourceDetailDataType
}

const RouterPanel = ({ data }: Props) => (
  <DetailSection icon={<FiInfo size={20} />} title="Router Details">
    <LayoutTwoCols>
      {data.status && (
        <ItemBlock title="Status">
          <ResourceStatus status={data.status} withLabel />
        </ItemBlock>
      )}
      {data.provider && (
        <ItemBlock title="Provider">
          <ProviderIcon name={data.provider} />
          <ProviderName css={{ ml: '$2' }}>{data.provider}</ProviderName>
        </ItemBlock>
      )}
      {data.priority && (
        <ItemBlock title="Priority">
          <Tooltip label={data.priority.toString()} action="copy">
            <Text css={{ overflowWrap: 'break-word' }}>{data.priority.toString()}</Text>
          </Tooltip>
        </ItemBlock>
      )}
    </LayoutTwoCols>
    {data.rule ? (
      <ItemBlock title="Rule">
        <Tooltip label={data.rule} action="copy">
          <Text css={{ overflowWrap: 'break-word' }}>{data.rule}</Text>
        </Tooltip>
      </ItemBlock>
    ) : null}
    {data.name && (
      <ItemBlock title="Name">
        <Tooltip label={data.name} action="copy">
          <Text css={{ overflowWrap: 'break-word' }}>{data.name}</Text>
        </Tooltip>
      </ItemBlock>
    )}
    {!!data.using && data.using && data.using.length > 0 && (
      <ItemBlock title="Entrypoints">
        {data.using.map((ep) => (
          <Tooltip key={ep} label={ep} action="copy">
            <Badge css={{ mr: '$2' }}>{ep}</Badge>
          </Tooltip>
        ))}
      </ItemBlock>
    )}
    {data.service && (
      <ItemBlock title="Service">
        <Tooltip label={data.service} action="copy">
          <Text css={{ overflowWrap: 'break-word' }}>{data.service}</Text>
        </Tooltip>
      </ItemBlock>
    )}
    {data.error && (
      <ItemBlock title="Errors">
        <GenericTable items={data.error} status="error" />
      </ItemBlock>
    )}
  </DetailSection>
)

export default RouterPanel
