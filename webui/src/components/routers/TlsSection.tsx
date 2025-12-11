import { Badge, Box, Card, Flex } from '@traefiklabs/faency'
import { useMemo } from 'react'

import TlsIcon from './TlsIcon'

import { EmptyIcon } from 'components/icons/EmptyIcon'
import { BooleanState, EmptyPlaceholder } from 'components/resources/DetailItemComponents'
import DetailsCard, { SectionTitle } from 'components/resources/DetailsCard'

type Props = {
  data?: Router.TLS
}

const TlsSection = ({ data }: Props) => {
  const items = useMemo(() => {
    if (data) {
      return [
        data?.options && { key: 'Options', val: data.options },
        { key: 'Passthrough', val: <BooleanState enabled={!!data.passthrough} /> },
        data?.certResolver && { key: 'Certificate resolver', val: data.certResolver },
        data?.domains && {
          stackVertical: true,
          forceNewRow: true,
          key: 'Domains',
          val: (
            <Flex css={{ flexDirection: 'column' }}>
              {data.domains?.map((domain) => (
                <Flex key={domain.main} css={{ flexWrap: 'wrap' }}>
                  <a href={`//${domain.main}`}>
                    <Badge variant="blue" css={{ mr: '$2', mb: '$2', color: '$primary', borderColor: '$primary' }}>
                      {domain.main}
                    </Badge>
                  </a>
                  {domain.sans?.map((sub) => (
                    <a key={sub} href={`//${sub}`}>
                      <Badge css={{ mr: '$2', mb: '$2' }}>{sub}</Badge>
                    </a>
                  ))}
                </Flex>
              ))}
            </Flex>
          ),
        },
      ].filter(Boolean) as { key: string; val: string | React.ReactElement }[]
    }
  }, [data])
  return (
    <Flex direction="column" gap={2}>
      <SectionTitle icon={<TlsIcon />} title="TLS" />
      {items?.length ? (
        <DetailsCard items={items} />
      ) : (
        <Card>
          <Flex direction="column" align="center" justify="center" css={{ flexGrow: 1, textAlign: 'center', py: '$4' }}>
            <Box
              css={{
                width: 56,
                svg: {
                  width: '100%',
                  height: '100%',
                },
              }}
            >
              <EmptyIcon />
            </Box>
            <EmptyPlaceholder css={{ mt: '$3' }}>
              There is no
              <br />
              TLS configured
            </EmptyPlaceholder>
          </Flex>
        </Card>
      )}
    </Flex>
  )
}

export default TlsSection
