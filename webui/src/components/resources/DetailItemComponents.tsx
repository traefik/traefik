import { Badge, CSS, Flex, styled, Text } from '@traefiklabs/faency'
import { ReactNode } from 'react'
import { BsToggleOff, BsToggleOn } from 'react-icons/bs'

import { colorByStatus } from './Status'

import CopyButton from 'components/buttons/CopyButton'

export const ItemTitle = styled(Text, {
  marginBottom: '$3',
  color: 'hsl(0, 0%, 56%)',
  fontSize: '12px',
  fontWeight: 600,
  textAlign: 'left',
  textTransform: 'capitalize',
  wordBreak: 'break-word',
})

const ItemBlockContainer = styled(Flex, {
  maxWidth: '100%',
  flexWrap: 'wrap !important',
  rowGap: '$2',

  // This forces the Tooltips to respect max-width, since we can't define
  // it directly on the component, and the Chips are automatically covered.
  span: {
    maxWidth: '100%',
  },
})

const FlexLimited = styled(Flex, {
  maxWidth: '100%',
  margin: '0 -8px -8px 0',
  span: {
    maxWidth: '100%',
  },
})

type ChipsType = {
  items: string[]
  variant?: 'gray' | 'red' | 'blue' | 'green' | 'neon' | 'orange' | 'purple'
  alignment?: 'center' | 'left'
}

export const Chips = ({ items, variant, alignment = 'left' }: ChipsType) => (
  <FlexLimited wrap="wrap">
    {items.map((item, index) => (
      <Badge key={index} variant={variant} css={{ textAlign: alignment, mr: '$2', mb: '$2' }}>
        <Flex gap={1} align="center">
          {item}
          <CopyButton text={item} iconOnly />
        </Flex>
      </Badge>
    ))}
  </FlexLimited>
)

type ItemBlockType = {
  title: string
  children?: ReactNode
}

export const ItemBlock = ({ title, children }: ItemBlockType) => (
  <Flex css={{ flexDirection: 'column', '&:not(:last-child)': { mb: '$5' } }}>
    <ItemTitle>{title}</ItemTitle>
    <ItemBlockContainer css={{ alignItems: 'center' }}>{children}</ItemBlockContainer>
  </Flex>
)

export const BooleanState = ({ enabled, css }: { enabled: boolean; css?: CSS }) => (
  <Flex align="center" gap={2} css={{ color: '$textDefault', ...css }}>
    {enabled ? (
      <BsToggleOn color={colorByStatus.enabled} size={24} data-testid={`enabled-true`} />
    ) : (
      <BsToggleOff color="inherit" size={24} data-testid={`enabled-false`} />
    )}

    <Text css={{ color: enabled ? colorByStatus.enabled : 'inherit', fontWeight: 600, fontSize: 'inherit' }}>
      {enabled ? 'True' : 'False'}
    </Text>
  </Flex>
)

export const ProviderName = styled(Text, {
  textTransform: 'capitalize',
  overflowWrap: 'break-word',
  fontSize: 'inherit !important',
})

export const EmptyPlaceholder = styled(Text, {
  color: 'hsl(0, 0%, 76%)',
  fontSize: '20px',
  fontWeight: '700',
  lineHeight: '1.2',
})
