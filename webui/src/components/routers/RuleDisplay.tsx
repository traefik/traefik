import { Badge, Box, Flex, Text } from '@traefik-labs/faency'
import CopyButton from 'components/buttons/CopyButton'
import Tooltip from 'components/Tooltip'
import { flattenMatchers, MatcherToken, OperatorToken, parseRule, Token } from 'libs/rule-parser'
import { ReactNode, useState } from 'react'

type BadgeVariant = 'gray' | 'red' | 'blue' | 'green' | 'orange'

const MATCHER_VARIANT: Record<string, BadgeVariant> = {
  Host: 'blue',
  HostRegexp: 'blue',
  ClientIP: 'orange',
  Path: 'green',
  PathPrefix: 'green',
  PathRegexp: 'green',
  Method: 'gray',
  Header: 'orange',
  HeaderRegexp: 'orange',
  Query: 'red',
  QueryRegexp: 'red',
}

const VARIANT_TONAL: Record<BadgeVariant, { bg: string; fg: string; border: string }> = {
  blue:   { bg: '$blue4',   fg: '$blue11',   border: '$blue6' },
  green:  { bg: '$green4',  fg: '$green11',  border: '$green6' },
  orange: { bg: '$orange4', fg: '$orange11', border: '$orange6' },
  red:    { bg: '$red4',    fg: '$red11',    border: '$red6' },
  gray:   { bg: '$gray4',   fg: '$gray11',   border: '$gray6' },
}

function matcherVariant(name: string): BadgeVariant {
  return MATCHER_VARIANT[name] ?? 'gray'
}

function tonalCss(variant: BadgeVariant) {
  const t = VARIANT_TONAL[variant]
  return { backgroundColor: t.bg, color: t.fg, borderColor: t.border }
}

function matcherParts(token: MatcherToken): { key: string; value: string } {
  const prefix = token.negated ? '!' : ''
  return { key: `${prefix}${token.name}`, value: token.args.join(', ') }
}

function matcherLabel(token: MatcherToken): string {
  const { key, value } = matcherParts(token)
  return `${key}: ${value}`
}

type MatcherSegment = { kind: 'matcher'; token: MatcherToken }
type OrBracketSegment = { kind: 'or-box'; segments: Segment[] }
type FallbackSegment = { kind: 'fallback'; raw: string }
type Segment = MatcherSegment | OrBracketSegment | FallbackSegment

// Does NOT recurse into toSegments to avoid producing a nested or-box inside an existing one.
function tokenToSimpleSegment(token: Token): Segment {
  if (token.type === 'matcher') return { kind: 'matcher', token }
  if (token.type === 'fallback') return { kind: 'fallback', raw: token.raw }
  if (token.type === 'group') {
    const matchers = flattenMatchers([token])
    if (matchers.length === 1) return { kind: 'matcher', token: matchers[0] }
    return { kind: 'or-box', segments: matchers.map((m) => ({ kind: 'matcher' as const, token: m })) }
  }
  return { kind: 'fallback', raw: '' }
}

function toSegments(tokens: Token[]): Segment[] {
  const segments: Segment[] = []
  let i = 0

  while (i < tokens.length) {
    const token = tokens[i]

    if (token.type === 'operator') {
      i++
      continue
    }

    if (token.type === 'group') {
      const hasOr = token.tokens.some(
        (t): t is OperatorToken => t.type === 'operator' && t.value === '||',
      )
      if (hasOr) {
        // Extract matchers directly — calling toSegments on group.tokens would create a nested or-box.
        const orItems: Segment[] = token.tokens
          .filter((t): t is MatcherToken => t.type === 'matcher')
          .map((m) => ({ kind: 'matcher' as const, token: m }))
        segments.push({ kind: 'or-box', segments: orItems })
      } else {
        segments.push(...toSegments(token.tokens))
      }
      i++
      continue
    }

    const nextIdx = i + 1
    if (
      nextIdx < tokens.length &&
      tokens[nextIdx].type === 'operator' &&
      (tokens[nextIdx] as OperatorToken).value === '||'
    ) {
      const orItems: Segment[] = [tokenToSimpleSegment(token)]
      let j = nextIdx
      while (
        j < tokens.length &&
        tokens[j].type === 'operator' &&
        (tokens[j] as OperatorToken).value === '||'
      ) {
        j++
        if (j < tokens.length) {
          orItems.push(tokenToSimpleSegment(tokens[j]))
          j++
        }
      }
      segments.push({ kind: 'or-box', segments: orItems })
      i = j
      continue
    }

    segments.push(tokenToSimpleSegment(token))
    i++
  }

  return segments
}

const MatcherPill = ({
  token,
  size,
  withCopy,
}: {
  token: MatcherToken
  size?: 'small' | 'large'
  withCopy?: boolean
}) => {
  const label = matcherLabel(token)
  const { key, value } = matcherParts(token)
  const variant = matcherVariant(token.name)
  return (
    <Badge
      variant={variant}
      size={size}
      css={{
        width: '100%',
        overflow: 'hidden',
        textAlign: 'left',
        justifyContent: 'flex-start',
        ...tonalCss(variant),
      }}
    >
      <Flex gap={1} align="center" css={{ overflow: 'hidden', flex: 1 }}>
        <Text
          css={{
            fontSize: size === 'small' ? '11px' : '12px',
            opacity: 0.75,
            fontWeight: 600,
            whiteSpace: 'nowrap',
            flexShrink: 0,
            color: 'inherit',
          }}
        >
          {key}:
        </Text>
        <Text
          css={{
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            fontSize: 'inherit',
            color: 'inherit',
            flex: 1,
            minWidth: 0,
          }}
        >
          {value}
        </Text>
        {withCopy && <CopyButton text={label} iconOnly />}
      </Flex>
    </Badge>
  )
}

const OrBracket = ({ children }: { children: ReactNode }) => (
  <Flex css={{ alignItems: 'stretch', gap: '$1', width: '100%' }}>
    <Flex direction="column" css={{ gap: '$1', flex: 1, minWidth: 0 }}>
      {children}
    </Flex>
    <Box
      css={{
        borderTop: '1px solid $gray10',
        borderRight: '1px solid $gray10',
        borderBottom: '1px solid $gray10',
        borderRadius: '0 4px 4px 0',
        width: 8,
        alignSelf: 'stretch',
        flexShrink: 0,
      }}
    />
    <Flex align="center" css={{ flexShrink: 0 }}>
      <Text css={{ fontSize: '11px', fontWeight: 700, color: '$gray11', whiteSpace: 'nowrap' }}>OR</Text>
    </Flex>
  </Flex>
)

function renderSegments(segments: Segment[], withCopy: boolean): ReactNode[] {
  return segments.map((seg, idx) => {
    if (seg.kind === 'fallback') {
      return (
        <Text key={idx} css={{ wordBreak: 'break-word', fontSize: '$3' }}>
          {seg.raw}
        </Text>
      )
    }
    if (seg.kind === 'matcher') {
      return (
        <Tooltip key={idx} label={matcherLabel(seg.token)}>
          <Box css={{ width: '100%' }}>
            <MatcherPill token={seg.token} withCopy={withCopy} />
          </Box>
        </Tooltip>
      )
    }
    if (seg.kind === 'or-box') {
      return <OrBracket key={idx}>{renderSegments(seg.segments, withCopy)}</OrBracket>
    }
    return null
  })
}

function buildCompactNodes(
  segments: Segment[],
  counter: { n: number },
  limit: number | null,
): ReactNode[] {
  const nodes: ReactNode[] = []

  for (const seg of segments) {
    if (limit !== null && counter.n >= limit) break

    if (seg.kind === 'fallback') {
      nodes.push(
        <Text key={nodes.length} css={{ wordBreak: 'break-word', fontSize: '$3' }}>
          {seg.raw}
        </Text>,
      )
    } else if (seg.kind === 'matcher') {
      nodes.push(
        <Tooltip key={nodes.length} label={matcherLabel(seg.token)}>
          <Box css={{ width: '100%' }}>
            <MatcherPill token={seg.token} size="small" withCopy />
          </Box>
        </Tooltip>,
      )
      counter.n++
    } else if (seg.kind === 'or-box') {
      const inner = buildCompactNodes(seg.segments, counter, limit)
      if (inner.length > 0) {
        nodes.push(<OrBracket key={nodes.length}>{inner}</OrBracket>)
      }
    }
  }

  return nodes
}

const COMPACT_LIMIT = 3

type RuleDisplayProps = {
  rule?: string
  compact?: boolean
}

export default function RuleDisplay({ rule, compact = false }: RuleDisplayProps) {
  const [expanded, setExpanded] = useState(false)

  if (!rule) return null

  const tokens = parseRule(rule)
  const segments = toSegments(tokens)

  if (compact) {
    const totalMatchers = flattenMatchers(tokens).length
    const hasOverflow = totalMatchers > COMPACT_LIMIT
    const limit = !expanded && hasOverflow ? COMPACT_LIMIT : null
    const nodes = buildCompactNodes(segments, { n: 0 }, limit)

    return (
      <Flex direction="column" css={{ gap: '$1', width: '100%' }}>
        {nodes}
        {hasOverflow && (
          <Badge
            variant="gray"
            size="small"
            interactive
            css={{ cursor: 'pointer', flexShrink: 0, ...tonalCss('gray') }}
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
              setExpanded((v) => !v)
            }}
          >
            {expanded ? 'show less' : `+${totalMatchers - COMPACT_LIMIT} more`}
          </Badge>
        )}
      </Flex>
    )
  }

  return (
    <Flex direction="column" css={{ gap: '$1', width: '100%' }}>
      {renderSegments(segments, false)}
    </Flex>
  )
}
