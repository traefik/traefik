import { Badge, Box, Flex, Text } from '@traefiklabs/faency'
import { FiShield } from 'react-icons/fi'

import { BooleanState, DetailSection, EmptyPlaceholder, ItemBlock } from './DetailSections'

import { EmptyIcon } from 'components/icons/EmptyIcon'
import { RouterDetailType } from 'hooks/use-resource-detail'

type Props = {
  data: RouterDetailType
}

const TlsPanel = ({ data }: Props) => (
  <DetailSection icon={<FiShield size={20} />} title="TLS">
    {data.tls ? (
      <Flex css={{ flexDirection: 'column' }}>
        <ItemBlock title="TLS">
          <BooleanState enabled />
        </ItemBlock>
        {data.tls.options && (
          <ItemBlock title="Options">
            <Text css={{ overflowWrap: 'break-word' }}>{data.tls.options}</Text>
          </ItemBlock>
        )}
        <ItemBlock title="PassThrough">
          <BooleanState enabled={!!data.tls.passthrough} />
        </ItemBlock>
        {data.tls.certResolver && (
          <ItemBlock title="Certificate Resolver">
            <Text css={{ overflowWrap: 'break-word' }}>{data.tls.certResolver}</Text>
          </ItemBlock>
        )}
        {data.tls.domains && (
          <ItemBlock title="Domains">
            <Flex css={{ flexDirection: 'column' }}>
              {data.tls.domains?.map((domain) => (
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
          </ItemBlock>
        )}
      </Flex>
    ) : (
      <Flex direction="column" align="center" justify="center" css={{ flexGrow: 1, textAlign: 'center' }}>
        <Box
          css={{
            width: 88,
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
    )}
  </DetailSection>
)

export default TlsPanel
