import { Box, Flex, H3, styled, Text } from '@traefiklabs/faency'
import { FiLayers } from 'react-icons/fi'

import { DetailSection, EmptyPlaceholder, ItemBlock, LayoutTwoCols, ProviderName } from './DetailSections'
import GenericTable from './GenericTable'
import { RenderUnknownProp } from './RenderUnknownProp'
import { ResourceStatus } from './ResourceStatus'

import { EmptyIcon } from 'components/icons/EmptyIcon'
import ProviderIcon from 'components/icons/providers'
import { Middleware, RouterDetailType } from 'hooks/use-resource-detail'
import { parseMiddlewareType } from 'libs/parsers'

const Separator = styled('hr', {
  border: 'none',
  background: '$tableRowBorder',
  margin: '0 0 24px',
  height: '1px',
  minHeight: '1px',
})

const filterMiddlewareProps = (middleware: Middleware): string[] => {
  const filteredProps = [] as string[]
  const propsToRemove = ['name', 'plugin', 'status', 'type', 'provider', 'error', 'usedBy', 'routers']

  Object.keys(middleware).map((propName) => {
    if (!propsToRemove.includes(propName)) {
      filteredProps.push(propName)
    }
  })

  return filteredProps
}

type RenderMiddlewareProps = {
  middleware: Middleware
  withHeader?: boolean
}

export const RenderMiddleware = ({ middleware, withHeader }: RenderMiddlewareProps) => (
  <Flex key={middleware.name} css={{ flexDirection: 'column' }}>
    {withHeader && <H3 css={{ mb: '$7', overflowWrap: 'break-word' }}>{middleware.name}</H3>}
    <LayoutTwoCols>
      {(middleware.type || middleware.plugin) && (
        <ItemBlock title="Type">
          <Text css={{ lineHeight: '32px', overflowWrap: 'break-word' }}>{parseMiddlewareType(middleware)}</Text>
        </ItemBlock>
      )}
      {middleware.provider && (
        <ItemBlock title="Provider">
          <ProviderIcon name={middleware.provider} />
          <ProviderName css={{ ml: '$2' }}>{middleware.provider}</ProviderName>
        </ItemBlock>
      )}
    </LayoutTwoCols>
    {middleware.status && (
      <ItemBlock title="Status">
        <ResourceStatus status={middleware.status} withLabel />
      </ItemBlock>
    )}
    {middleware.error && (
      <ItemBlock title="Errors">
        <GenericTable items={middleware.error} status="error" />
      </ItemBlock>
    )}
    {middleware.plugin &&
      Object.keys(middleware.plugin).map((pluginName) => (
        <RenderUnknownProp key={pluginName} name={pluginName} prop={middleware.plugin?.[pluginName]} />
      ))}
    {filterMiddlewareProps(middleware).map((propName) => (
      <RenderUnknownProp
        key={propName}
        name={propName}
        prop={middleware[propName]}
        removeTitlePrefix={middleware.type}
      />
    ))}
  </Flex>
)

const MiddlewarePanel = ({ data }: { data: RouterDetailType }) => (
  <DetailSection icon={<FiLayers size={20} />} title="Middlewares">
    {data.middlewares ? (
      data.middlewares.map((middleware, index) => (
        <Box key={middleware.name}>
          <RenderMiddleware middleware={middleware} withHeader />
          {data.middlewares && index < data.middlewares.length - 1 && <Separator />}
        </Box>
      ))
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
          There are no
          <br />
          Middlewares configured
        </EmptyPlaceholder>
      </Flex>
    )}
  </DetailSection>
)

export default MiddlewarePanel
