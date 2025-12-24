import { AriaTable, AriaTbody, AriaTd, AriaTr, Flex, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'

import Status from './Status'

import CopyableText from 'components/CopyableText'

type GenericTableProps = {
  items: (number | string)[]
  status?: Resource.Status
  copyable?: boolean
}

export default function GenericTable({ items, status, copyable = false }: GenericTableProps) {
  const border = useMemo(() => `1px solid $${status === 'error' ? 'textRed' : 'tableRowBorder'}`, [status])

  return (
    <AriaTable css={{ wordBreak: 'break-word', boxShadow: 'none', border }}>
      <AriaTbody>
        {items.map((item, index) => (
          <AriaTr key={index}>
            <AriaTd css={{ p: '$2' }}>
              <Flex align="start" gap={2} css={{ width: 'fit-content' }}>
                {status ? (
                  <Status status="error" css={{ p: '4px', marginRight: 0 }} size={14} />
                ) : (
                  <Text css={{ fontFamily: 'monospace', mt: '1px', userSelect: 'none' }} variant="subtle">
                    {index}
                  </Text>
                )}
                {copyable ? (
                  <CopyableText
                    text={String(item)}
                    css={{
                      fontFamily: status === 'error' ? 'monospace' : undefined,
                      color: status === 'error' ? '$textRed' : 'initial',
                    }}
                  />
                ) : (
                  <Text
                    css={{ fontFamily: status === 'error' ? 'monospace' : undefined }}
                    variant={status === 'error' ? 'red' : undefined}
                  >
                    {item}
                  </Text>
                )}
              </Flex>
            </AriaTd>
          </AriaTr>
        ))}
      </AriaTbody>
    </AriaTable>
  )
}
