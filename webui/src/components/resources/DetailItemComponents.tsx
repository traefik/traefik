import { Badge, CSS, Flex, styled, Text } from '@traefik-labs/faency'
import { ReactNode, useState } from 'react'
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
  limit?: number
}

export const Chips = ({ items, variant, alignment = 'left', limit }: ChipsType) => {
  const [expanded, setExpanded] = useState(false)

  const hasOverflow = limit !== undefined && items.length > limit
  const visible = hasOverflow && !expanded ? items.slice(0, limit) : items

  return (
    <Flex wrap={expanded ? 'wrap' : 'nowrap'} css={{ gap: '$2', maxWidth: '100%', minWidth: 0, overflow: 'hidden' }}>
      {visible.map((item, index) => (
        <Badge key={index} variant={variant} css={{ textAlign: alignment, flexShrink: 1, minWidth: '48px', overflow: 'hidden' }}>
          <Flex gap={1} align="center" css={{ minWidth: 0 }}>
            <Text css={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: 'inherit', color: 'inherit', minWidth: 0 }}>
              {item}
            </Text>
            <CopyButton text={item} iconOnly />
          </Flex>
        </Badge>
      ))}
      {hasOverflow && (
        <Badge
          variant="gray"
          interactive
          css={{
            cursor: 'pointer',
            flexShrink: 0,
            background: 'transparent',
            border: '1px dashed $gray8',
            color: '$gray11',
          }}
          onClick={(e) => {
            e.preventDefault()
            e.stopPropagation()
            setExpanded((v) => !v)
          }}
        >
          {expanded ? 'show less' : `+${items.length - limit!} more`}
        </Badge>
      )}
    </Flex>
  )
}

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
