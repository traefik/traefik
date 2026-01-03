import { Card, CSS, Flex, Grid, H2, Skeleton, styled, Text } from '@traefiklabs/faency'
import { Fragment, ReactNode, useMemo } from 'react'

import ScrollableCard from 'components/ScrollableCard'
import breakpoints from 'utils/breakpoints'

const StyledText = styled(Text, {
  fontSize: 'inherit !important',
  lineHeight: '24px',
})

export const ValText = styled(StyledText, {
  overflowWrap: 'break-word',
  wordBreak: 'break-word',
})

export const SectionTitle = ({ icon, title }: { icon?: ReactNode; title: string }) => {
  return (
    <Flex gap={2} align="center" css={{ color: '$headingDefault' }}>
      {icon && icon}
      <H2 css={{ fontSize: '$5' }}>{title}</H2>
    </Flex>
  )
}

type DetailsCardProps = {
  css?: CSS
  keyColumns?: number
  items: { key: string; val: string | React.ReactElement; stackVertical?: boolean; forceNewRow?: boolean }[]
  minKeyWidth?: string
  maxKeyWidth?: string
  testidPrefix?: string
  testId?: string
  title?: string
  icon?: ReactNode
  scrollable?: boolean
}

export default function DetailsCard({
  css = {},
  keyColumns = 2,
  items,
  minKeyWidth,
  maxKeyWidth,
  testidPrefix = 'definition',
  testId,
  title,
  icon,
  scrollable = false,
}: DetailsCardProps) {
  const ParentComponent = useMemo(() => {
    if (scrollable) {
      return ScrollableCard
    }
    return Card
  }, [scrollable])

  return (
    <Flex as="section" direction="column" gap={2} css={{ ...css }} data-testid={testId || `${testidPrefix}-section`}>
      {title ? <SectionTitle icon={icon} title={title} /> : null}
      <ParentComponent css={{ flex: 1 }}>
        <Grid
          css={{
            gap: '$2 $3',
            gridTemplateColumns: maxKeyWidth
              ? `repeat(${keyColumns}, minmax(auto, ${maxKeyWidth}) 1fr)`
              : `repeat(${keyColumns}, auto 1fr)`,
            [`@media (max-width:${breakpoints.laptop}px)`]: {
              gridTemplateColumns: maxKeyWidth ? `minmax(auto, ${maxKeyWidth}) 1fr` : 'auto 1fr',
            },
          }}
        >
          {items.map((item, index) => {
            // Handle forceNewRow props
            const cellsBeforeThis = items.slice(0, index).reduce((count, prevItem) => {
              if (prevItem.stackVertical) return count + keyColumns
              return count + 1
            }, 0)

            const needsEmptyCell = item.forceNewRow && cellsBeforeThis % keyColumns !== 0

            return (
              <Fragment key={index}>
                {needsEmptyCell && (
                  <>
                    <div />
                    <div />
                  </>
                )}
                {item.stackVertical ? (
                  <Flex direction="column" gap={2} css={{ gridColumn: 'span 2' }}>
                    <StyledText
                      css={{
                        fontWeight: 600,
                        minWidth: minKeyWidth,
                        maxWidth: maxKeyWidth,
                        overflowWrap: 'break-word',
                        wordBreak: 'break-word',
                      }}
                    >
                      {item.key}
                    </StyledText>
                    {typeof item.val === 'string' ? (
                      <ValText>{item.val}</ValText>
                    ) : (
                      <Flex
                        css={{
                          '> *': {
                            height: 'fit-content',
                          },
                          height: '100%',
                        }}
                      >
                        {item.val}
                      </Flex>
                    )}
                  </Flex>
                ) : (
                  <>
                    <Grid>
                      {index < keyColumns
                        ? items
                            .filter((hiddenItem) => hiddenItem.key != item.key)
                            .map((hiddenItem, jndex) => (
                              <StyledText
                                key={`hidden-${index}-${jndex}`}
                                aria-hidden="true"
                                css={{
                                  gridArea: '1 / 1',
                                  fontWeight: 600,
                                  visibility: 'hidden',
                                  maxWidth: maxKeyWidth,
                                }}
                              >
                                {hiddenItem.key}
                              </StyledText>
                            ))
                        : null}
                      <StyledText
                        css={{
                          gridArea: '1 / 1',
                          fontWeight: 600,
                          minWidth: minKeyWidth,
                          maxWidth: maxKeyWidth,
                          overflowWrap: 'break-word',
                          wordBreak: 'break-word',
                        }}
                      >
                        {item.key}
                      </StyledText>
                    </Grid>
                    {typeof item.val === 'string' ? (
                      <ValText css={{ flex: 1 }}>{item.val}</ValText>
                    ) : (
                      <Flex
                        align="center"
                        css={{
                          alignSelf: 'start',
                          '> *': {
                            height: 'fit-content',
                          },
                          height: '100%',
                        }}
                      >
                        {item.val}
                      </Flex>
                    )}
                  </>
                )}
              </Fragment>
            )
          })}
        </Grid>
      </ParentComponent>
    </Flex>
  )
}

export function DetailsCardSkeleton({
  keyColumns = 2,
  rows = 3,
  testidPrefix = 'definition',
  title,
  icon,
}: { rows?: number } & Omit<DetailsCardProps, 'items'>) {
  return (
    <Flex as="section" direction="column" gap={2} data-testid={`${testidPrefix}-section-skeleton`}>
      {title ? <SectionTitle icon={icon} title={title} /> : <Skeleton css={{ height: '$5', width: '150px' }} />}
      <Card css={{ flex: 1 }}>
        <Grid
          css={{
            gap: '$2 $3',
            gridTemplateColumns: `repeat(${keyColumns}, auto 1fr)`,
            [`@media (max-width:${breakpoints.laptop}px)`]: { gridTemplateColumns: 'auto 1fr' },
          }}
        >
          {[...Array(rows * keyColumns)].map((_, idx) => (
            <Fragment key={idx}>
              <Skeleton css={{ height: '$5', width: '96px' }} />
              <Skeleton css={{ height: '$5', width: '192px' }} />
            </Fragment>
          ))}
        </Grid>
      </Card>
    </Flex>
  )
}
