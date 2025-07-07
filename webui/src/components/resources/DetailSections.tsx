import { Badge, Box, Card, Flex, H2, styled, Text } from '@traefiklabs/faency'
import { ReactNode } from 'react'
import { FiArrowRight, FiToggleLeft, FiToggleRight } from 'react-icons/fi'
import { useNavigate } from 'react-router-dom'

import { StatusWrapper } from './ResourceStatus'
import { colorByStatus } from './Status'

import Tooltip from 'components/Tooltip'

const CustomHeading = styled(H2, {
  display: 'flex',
  alignItems: 'center',
})

type SectionHeaderType = {
  icon?: ReactNode
  title?: string | undefined
}

export const SectionHeader = ({ icon, title }: SectionHeaderType) => {
  if (!title) {
    return (
      <CustomHeading css={{ mb: '$6' }}>
        <Box css={{ width: 5, height: 4, bg: 'hsl(220, 6%, 90%)', borderRadius: 1 }} />
        <Box css={{ width: '50%', maxWidth: '300px', height: 4, bg: 'hsl(220, 6%, 90%)', borderRadius: 1, ml: '$2' }} />
      </CustomHeading>
    )
  }

  return (
    <CustomHeading css={{ mb: '$5' }}>
      {icon ? icon : null}
      <Text size={6} css={{ ml: '$2' }}>
        {title}
      </Text>
    </CustomHeading>
  )
}

export const ItemTitle = styled(Text, {
  marginBottom: '$3',
  color: 'hsl(0, 0%, 56%)',
  letterSpacing: '3px',
  fontSize: '12px',
  fontWeight: 600,
  textAlign: 'left',
  textTransform: 'uppercase',
  wordBreak: 'break-word',
})

const SpacedCard = styled(Card, {
  '& + &': {
    marginTop: '16px',
  },
})

const CardDescription = styled(Text, {
  textAlign: 'left',
  fontWeight: '700',
  fontSize: '16px',
  lineHeight: '16px',
  wordBreak: 'break-word',
})

const CardListColumnWrapper = styled(Flex, {
  display: 'flex',
})

const CardListColumn = styled(Flex, {
  minWidth: 160,
  maxWidth: '66%',
  maxHeight: '416px',
  overflowY: 'auto',
  p: '$1',
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

const FlexLink = styled('a', {
  display: 'flex',
  flexFlow: 'column',
  textDecoration: 'none',
})

type CardType = {
  title: string
  description?: string
  focus?: boolean
  link?: string
}

type SectionType = SectionHeaderType & {
  cards?: CardType[] | undefined
  isLast?: boolean
  bigDescription?: boolean
}
const CardSkeleton = ({ bigDescription }: { bigDescription?: boolean }) => {
  return (
    <SpacedCard css={{ p: '$3' }}>
      <ItemTitle>
        <Box css={{ height: '12px', bg: '$slate5', borderRadius: 1, mb: '$3', mr: '60%' }} />
      </ItemTitle>
      <CardDescription>
        <Box
          css={{
            height: bigDescription ? '22px' : '14px',
            mr: '20%',
            bg: '$slate5',
            borderRadius: 1,
          }}
        />
      </CardDescription>
    </SpacedCard>
  )
}

export const CardListSection = ({ icon, title, cards, isLast, bigDescription }: SectionType) => {
  const navigate = useNavigate()

  return (
    <Flex css={{ flexDirection: 'column', flexGrow: 1 }}>
      <SectionHeader icon={icon} title={title} />
      <CardListColumnWrapper>
        <CardListColumn>
          <Flex css={{ flexDirection: 'column', flexGrow: 1, marginRight: '$3' }}>
            {!cards && <CardSkeleton bigDescription={bigDescription} />}
            {cards
              ?.filter((c) => !!c.description)
              .map((card) => (
                <SpacedCard key={card.description} css={{ border: card.focus ? `2px solid $primary` : '', p: '$3' }}>
                  <FlexLink
                    data-testid={card.link}
                    onClick={(): false | void => !!card.link && navigate(card.link)}
                    css={{ cursor: card.link ? 'pointer' : 'inherit' }}
                  >
                    <ItemTitle>{card.title}</ItemTitle>
                    <CardDescription>{card.description}</CardDescription>
                  </FlexLink>
                </SpacedCard>
              ))}
            <Box css={{ height: '16px' }}>&nbsp;</Box>
          </Flex>
        </CardListColumn>
        {!isLast && (
          <Flex css={{ mt: '$5', mx: 'auto' }}>
            <FiArrowRight color="hsl(0, 0%, 76%)" size={24} />
          </Flex>
        )}
      </CardListColumnWrapper>
    </Flex>
  )
}

const FlexCard = styled(Card, {
  display: 'flex',
  flexFlow: 'column',
  flexGrow: '1',
  overflowY: 'auto',
  height: '600px',
})

const NarrowFlexCard = styled(FlexCard, {
  height: '400px',
})

const ItemTitleSkeleton = styled(Box, {
  height: '16px',
  backgroundColor: '$slate5',
  borderRadius: '3px',
})

const ItemDescriptionSkeleton = styled(Box, {
  height: '16px',
  backgroundColor: '$slate5',
  borderRadius: '3px',
})

type DetailSectionSkeletonType = {
  narrow?: boolean
}

export const DetailSectionSkeleton = ({ narrow }: DetailSectionSkeletonType) => {
  const Card = narrow ? NarrowFlexCard : FlexCard

  return (
    <Flex css={{ flexDirection: 'column' }}>
      <SectionHeader />
      <Card css={{ p: '$5' }}>
        <LayoutTwoCols css={{ mb: '$2' }}>
          <ItemTitleSkeleton css={{ width: '40%' }} />
          <ItemTitleSkeleton css={{ width: '40%' }} />
        </LayoutTwoCols>
        <LayoutTwoCols css={{ mb: '$5' }}>
          <ItemDescriptionSkeleton css={{ width: '90%' }} />
          <ItemDescriptionSkeleton css={{ width: '90%' }} />
        </LayoutTwoCols>
        <Flex css={{ mb: '$2' }}>
          <ItemTitleSkeleton css={{ width: '30%' }} />
        </Flex>
        <Flex css={{ mb: '$5' }}>
          <ItemDescriptionSkeleton css={{ width: '50%' }} />
        </Flex>
        <Flex css={{ mb: '$2' }}>
          <ItemTitleSkeleton css={{ width: '30%' }} />
        </Flex>
        <Flex css={{ mb: '$5' }}>
          <ItemDescriptionSkeleton css={{ width: '70%' }} />
        </Flex>
        <Flex css={{ mb: '$2' }}>
          <ItemTitleSkeleton css={{ width: '30%' }} />
        </Flex>
        <Flex css={{ mb: '$5' }}>
          <ItemDescriptionSkeleton css={{ width: '50%' }} />
        </Flex>
        <LayoutTwoCols css={{ mb: '$2' }}>
          <ItemTitleSkeleton css={{ width: '40%' }} />
          <ItemTitleSkeleton css={{ width: '40%' }} />
        </LayoutTwoCols>
        <LayoutTwoCols css={{ mb: '$5' }}>
          <ItemDescriptionSkeleton css={{ width: '90%' }} />
          <ItemDescriptionSkeleton css={{ width: '90%' }} />
        </LayoutTwoCols>
      </Card>
    </Flex>
  )
}

type DetailSectionType = SectionHeaderType & {
  children?: ReactNode
  noPadding?: boolean
  narrow?: boolean
}

export const DetailSection = ({ icon, title, children, narrow, noPadding }: DetailSectionType) => {
  const Card = narrow ? NarrowFlexCard : FlexCard

  return (
    <Flex css={{ flexDirection: 'column' }}>
      <SectionHeader icon={icon} title={title} />
      <Card css={{ padding: noPadding ? 0 : '$5' }}>{children}</Card>
    </Flex>
  )
}

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
      <Tooltip key={index} label={item} action="copy">
        <Badge variant={variant} css={{ textAlign: alignment, mr: '$2', mb: '$2' }}>
          {item}
        </Badge>
      </Tooltip>
    ))}
  </FlexLimited>
)

type ChipPropsListType = {
  data: {
    [key: string]: string
  }
  variant?: 'gray' | 'red' | 'blue' | 'green' | 'neon' | 'orange' | 'purple'
}

export const ChipPropsList = ({ data, variant }: ChipPropsListType) => (
  <Flex css={{ flexWrap: 'wrap' }}>
    {Object.entries(data).map((entry: [string, string]) => (
      <Badge key={entry[0]} variant={variant} css={{ textAlign: 'left', mr: '$2', mb: '$2' }}>
        {entry[1]}
      </Badge>
    ))}
  </Flex>
)

type ItemBlockType = {
  title: string
  children?: ReactNode
}

export const ItemBlock = ({ title, children }: ItemBlockType) => (
  <Flex css={{ flexDirection: 'column', mb: '$5' }}>
    <ItemTitle>{title}</ItemTitle>
    <ItemBlockContainer css={{ alignItems: 'center' }}>{children}</ItemBlockContainer>
  </Flex>
)

const LayoutCols = styled(Box, {
  display: 'grid',
  gridGap: '16px',
})

export const LayoutTwoCols = styled(LayoutCols, {
  gridTemplateColumns: 'repeat(2, minmax(50%, 1fr))',
})

export const LayoutThreeCols = styled(LayoutCols, {
  gridTemplateColumns: 'repeat(3, minmax(30%, 1fr))',
})

export const BooleanState = ({ enabled }: { enabled: boolean }) => (
  <Flex align="center" gap={2}>
    <StatusWrapper
      css={{
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: enabled ? colorByStatus.enabled : colorByStatus.disabled,
      }}
      data-testid={`enabled-${enabled}`}
    >
      {enabled ? <FiToggleRight color="#fff" size={20} /> : <FiToggleLeft color="#fff" size={20} />}
    </StatusWrapper>
    <Text css={{ color: enabled ? colorByStatus.enabled : colorByStatus.disabled, fontWeight: 600 }}>
      {enabled ? 'True' : 'False'}
    </Text>
  </Flex>
)

export const ProviderName = styled(Text, {
  textTransform: 'capitalize',
  overflowWrap: 'break-word',
})

export const EmptyPlaceholder = styled(Text, {
  color: 'hsl(0, 0%, 76%)',
  fontSize: '20px',
  fontWeight: '700',
  lineHeight: '1.2',
})
