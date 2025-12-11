import { Card, Flex, styled, Link, Tooltip, Box, Text, Skeleton } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { FiArrowRight, FiGlobe, FiLayers, FiLogIn, FiZap } from 'react-icons/fi'

import CopyableText from 'components/CopyableText'
import ProviderIcon from 'components/icons/providers'
import { ProviderName } from 'components/resources/DetailItemComponents'
import DetailsCard, { SectionTitle } from 'components/resources/DetailsCard'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import ScrollableCard from 'components/ScrollableCard'
import { useHrefWithReturnTo } from 'hooks/use-href-with-return-to'
import { useResourceDetail } from 'hooks/use-resource-detail'

const FlexContainer = styled(Flex, {
  gap: '$3',
  flexDirection: 'column !important',
  alignItems: 'center !important',
  flex: '1 1 0',
  minWidth: '0',
  maxWidth: '100%',
})

const ArrowSeparator = () => {
  return (
    <Flex css={{ color: '$textSubtle' }}>
      <FiArrowRight size={20} />
    </Flex>
  )
}

const LinkedNameAndStatus = ({ data }: { data: { status: Resource.Status; name: string; href?: string } }) => {
  const hrefWithReturnTo = useHrefWithReturnTo(data?.href || '')

  if (!data.href) {
    return (
      <Flex gap={2} css={{ minWidth: 0, flex: 1 }}>
        <Tooltip content="Service not found">
          <Box>
            <ResourceStatus status={data.status} />
          </Box>
        </Tooltip>

        <Text
          css={{
            wordBreak: 'break-word',
            overflowWrap: 'anywhere',
            flex: 1,
            minWidth: 0,
            fontSize: '$4',
          }}
        >
          {data.name}
        </Text>
      </Flex>
    )
  }
  return (
    <Flex gap={2} css={{ minWidth: 0, flex: 1 }}>
      <ResourceStatus status={data.status} />
      <Link
        data-testid={data.href}
        href={hrefWithReturnTo}
        css={{
          wordBreak: 'break-word',
          overflowWrap: 'anywhere',
          flex: 1,
          minWidth: 0,
        }}
      >
        {data.name}
      </Link>
    </Flex>
  )
}

type RouterFlowDiagramProps = {
  data: Resource.DetailsData
  protocol: 'http' | 'tcp' | 'udp'
}

const RouterFlowDiagram = ({ data, protocol }: RouterFlowDiagramProps) => {
  const displayedEntrypoints = useMemo(() => {
    return data?.entryPointsData?.map((point) => {
      if (!point.message) {
        return { key: point.name, val: point.address }
      } else {
        return { key: point.message, val: '' }
      }
    })
  }, [data?.entryPointsData])

  const routerDetailsItems = useMemo(
    () =>
      [
        data.status && { key: 'Status', val: <ResourceStatus status={data.status} withLabel /> },
        data.provider && {
          key: 'Provider',
          val: (
            <>
              <ProviderIcon name={data.provider} />
              <ProviderName css={{ ml: '$2' }}>{data.provider}</ProviderName>
            </>
          ),
        },
        data.priority && { key: 'Priority', val: data.priority },
        data.rule && { key: 'Rule', val: <CopyableText css={{ lineHeight: 1.2 }} text={data.rule} /> },
      ].filter(Boolean) as { key: string; val: string | React.ReactElement }[],
    [data.priority, data.provider, data.rule, data.status],
  )

  const serviceSlug = data.service?.includes('@')
    ? data.service
    : `${data.service ?? 'unknown'}@${data.provider ?? 'unknown'}`

  const { data: serviceData, error: serviceDataError } = useResourceDetail(serviceSlug ?? '', 'services')

  return (
    <Flex gap={2} data-testid="router-structure">
      {!!data.using?.length && (
        <>
          <FlexContainer>
            <SectionTitle icon={<FiLogIn size={20} />} title="Entrypoints" />
            {displayedEntrypoints?.length ? (
              <DetailsCard
                css={{ width: '100%' }}
                items={displayedEntrypoints}
                keyColumns={1}
                maxKeyWidth="70%"
                scrollable
              />
            ) : (
              <DiagramCardSkeleton />
            )}
          </FlexContainer>

          <ArrowSeparator />
        </>
      )}

      <FlexContainer data-testid="router-details">
        <SectionTitle icon={<FiGlobe size={20} />} title={`${protocol.toUpperCase()} Router`} />
        <DetailsCard css={{ width: '100%' }} items={routerDetailsItems} keyColumns={1} scrollable />
      </FlexContainer>

      {data.hasValidMiddlewares && (
        <>
          <ArrowSeparator />
          <FlexContainer>
            <SectionTitle icon={<FiLayers size={20} />} title={`${protocol.toUpperCase()} Middlewares`} />
            {data.middlewares ? (
              <ScrollableCard>
                <Flex direction="column" gap={3}>
                  {data.middlewares.map((mw, idx) => {
                    const data = {
                      name: mw.name,
                      status: mw.status,
                      href: `/${protocol}/middlewares/${mw.name}`,
                    }
                    return <LinkedNameAndStatus key={`mw-${idx}`} data={data} />
                  })}
                </Flex>
              </ScrollableCard>
            ) : (
              <DiagramCardSkeleton />
            )}
          </FlexContainer>
        </>
      )}

      <ArrowSeparator />

      <FlexContainer>
        <SectionTitle icon={<FiZap size={20} />} title="Service" />
        <Card css={{ width: '100%' }}>
          <LinkedNameAndStatus
            data={{
              name: data.service as string,
              status: !serviceDataError ? (serviceData?.status ?? 'loading') : 'disabled',
              href: !serviceDataError ? `/${protocol}/services/${serviceSlug}` : undefined,
            }}
          />
        </Card>
      </FlexContainer>
    </Flex>
  )
}

const DiagramCardSkeleton = () => {
  return (
    <Card css={{ width: '100%', height: 200, gap: '$3', display: 'flex', flexDirection: 'column' }}>
      {[...Array(5)].map((_, idx) => (
        <Skeleton key={`1-${idx}`} />
      ))}
    </Card>
  )
}

export const RouterFlowDiagramSkeleton = () => {
  return (
    <Flex gap={4}>
      {[...Array(4)].map((_, index) => [
        <FlexContainer key={`container-${index}`}>
          <Skeleton css={{ width: 100 }} />
          <DiagramCardSkeleton />
        </FlexContainer>,
        index < 3 && <ArrowSeparator key={`separator-${index}`} />,
      ])}
    </Flex>
  )
}

export default RouterFlowDiagram
