import { AriaTable, AriaTbody, AriaTd, AriaTr, Flex, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'

import Status, { StatusType } from './Status'

import Tooltip from 'components/Tooltip'

type GenericTableProps = {
  items: (number | string)[]
  status?: StatusType
}

export default function GenericTable({ items, status }: GenericTableProps) {
  const border = useMemo(() => `1px solid $${status === 'error' ? 'textRed' : 'tableRowBorder'}`, [status])

  return (
    <AriaTable css={{ wordBreak: 'break-word', boxShadow: 'none', border }}>
      <AriaTbody>
        {items.map((item, index) => (
          <AriaTr key={index}>
            <AriaTd css={{ p: '$2' }}>
              <Tooltip label={item.toString()} action="copy">
                <Flex align="start" gap={2} css={{ width: 'fit-content' }}>
                  {status ? (
                    <Status status="error" css={{ p: '4px', marginRight: 0 }} size={16} />
                  ) : (
                    <Text css={{ fontFamily: 'monospace', mt: '1px', userSelect: 'none' }} variant="subtle">
                      {index}
                    </Text>
                  )}
                  <Text
                    css={{ fontFamily: status === 'error' ? 'monospace' : undefined }}
                    variant={status === 'error' ? 'red' : undefined}
                  >
                    {item}
                  </Text>
                </Flex>
              </Tooltip>
            </AriaTd>
          </AriaTr>
        ))}
      </AriaTbody>
    </AriaTable>
  )
}
