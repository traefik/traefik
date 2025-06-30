import { AriaTable, AriaTbody, AriaTd, AriaTr, Badge, Flex, Text } from '@traefiklabs/faency'

import Tooltip from 'components/Tooltip'

export type IpStrategy = {
  depth: number
  excludedIPs: string[]
}

export default function IpStrategyTable({ ipStrategy }: { ipStrategy: IpStrategy }) {
  return (
    <AriaTable css={{ wordBreak: 'break-word', boxShadow: 'none', border: '1px solid $tableRowBorder' }}>
      <AriaTbody>
        {ipStrategy.depth ? (
          <AriaTr>
            <AriaTd css={{ width: '104px', p: '$2' }}>
              <Text variant="subtle">Depth</Text>
            </AriaTd>
            <AriaTd css={{ p: '$2' }}>
              <Tooltip label={ipStrategy.depth.toString()} action="copy">
                <Text>{ipStrategy.depth}</Text>
              </Tooltip>
            </AriaTd>
          </AriaTr>
        ) : null}
        {ipStrategy.excludedIPs ? (
          <AriaTr>
            <AriaTd css={{ width: '104px', p: '$2', verticalAlign: 'baseline' }}>
              <Text variant="subtle">Excluded IPs</Text>
            </AriaTd>
            <AriaTd css={{ p: '$2' }}>
              <Flex gap={1} css={{ flexWrap: 'wrap' }}>
                {ipStrategy.excludedIPs.map((ip, index) => (
                  <Tooltip key={index} label={ip} action="copy">
                    <Badge> {ip}</Badge>
                  </Tooltip>
                ))}
              </Flex>
            </AriaTd>
          </AriaTr>
        ) : null}
      </AriaTbody>
    </AriaTable>
  )
}
