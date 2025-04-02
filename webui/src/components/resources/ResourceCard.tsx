import { Card, CSS, Flex, Text } from '@traefiklabs/faency'
import { ReactNode } from 'react'

type ResourceCardProps = {
  children: ReactNode
  css?: CSS
  title?: string
  titleCSS?: CSS
}

const ResourceCard = ({ children, css, title, titleCSS = {} }: ResourceCardProps) => {
  return (
    <Card css={css}>
      <Flex direction="column" align="center" justify="center" gap={3} css={{ height: '100%', p: '$2' }}>
        {title && (
          <Text variant="subtle" css={{ letterSpacing: 3, fontSize: '$2', wordBreak: 'break-all', ...titleCSS }}>
            {title.toUpperCase()}
          </Text>
        )}
        {children}
      </Flex>
    </Card>
  )
}

export default ResourceCard
