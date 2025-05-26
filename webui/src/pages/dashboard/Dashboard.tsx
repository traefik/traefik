import { Card, CSS, Flex, Grid, H2, Text } from '@traefiklabs/faency'
import { ReactNode, useMemo } from 'react'
import useSWR from 'swr'

import ProviderIcon from 'components/icons/providers'
import FeatureCard, { FeatureCardSkeleton } from 'components/resources/FeatureCard'
import ResourceCard from 'components/resources/ResourceCard'
import TraefikResourceStatsCard, { StatsCardSkeleton } from 'components/resources/TraefikResourceStatsCard'
import Page from 'layout/Page'
import { capitalizeFirstLetter } from 'utils/string'

const RESOURCES = ['routers', 'services', 'middlewares']

const SectionContainer = ({
  title,
  children,
  childrenContainerCss,
  css,
}: {
  title: string
  children: ReactNode
  childrenContainerCss?: CSS
  css?: CSS
}) => {
  return (
    <Flex direction="column" gap={4} css={{ mt: '$4', ...css }}>
      <Flex align="center" gap={2} css={{ color: '$headingDefault', mb: '$4' }}>
        <H2 css={{ fontSize: '$8' }}>{title}</H2>
      </Flex>
      <Grid
        gap={6}
        css={{
          gridTemplateColumns: 'repeat(auto-fill, minmax(215px, 1fr))',
          alignItems: 'stretch',
          ...childrenContainerCss,
        }}
      >
        {children}
      </Grid>
    </Flex>
  )
}

type ResourceData = {
  errors: number
  warnings: number
  total: number
}

export const Dashboard = () => {
  const { data: entrypoints } = useSWR('/entrypoints')
  const { data: overview } = useSWR('/overview')

  const features = useMemo(
    () =>
      overview?.features
        ? Object.keys(overview?.features).map((key: string) => {
            return { name: key, value: overview.features[key] }
          })
        : [],
    [overview?.features],
  )

  const hasResources = useMemo(() => {
    const filterFn = (x: ResourceData) => !x.errors && !x.total && !x.warnings
    return {
      http: Object.values<ResourceData>(overview?.http || {}).filter(filterFn).length !== 3,
      tcp: Object.values<ResourceData>(overview?.tcp || {}).filter(filterFn).length !== 3,
      udp: Object.values<ResourceData>(overview?.udp || {}).filter(filterFn).length !== 2,
    }
  }, [overview])

  // @FIXME skeleton not correctly displayed if only using suspense
  if (!entrypoints || !overview) {
    return <DashboardSkeleton />
  }

  return (
    <Page title="Dashboard">
      <Flex direction="column" gap={6}>
        <SectionContainer title="Entrypoints" css={{ mt: 0 }}>
          {entrypoints?.map((i, idx) => (
            <ResourceCard
              key={`entrypoint-${i.name}-${idx}`}
              css={{
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                minHeight: '125px',
              }}
              title={i.name}
              titleCSS={{ textAlign: 'center' }}
            >
              <Text css={{ fontSize: '$11', fontWeight: 500, wordBreak: 'break-word' }}>{i.address}</Text>
            </ResourceCard>
          ))}
        </SectionContainer>

        <SectionContainer
          title="HTTP"
          childrenContainerCss={{ gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))' }}
        >
          {overview?.http && hasResources.http ? (
            RESOURCES.map((i) => (
              <TraefikResourceStatsCard
                key={`http-${i}`}
                title={capitalizeFirstLetter(i)}
                data-testid={`section-http-${i}`}
                linkTo={`/http/${i}`}
                {...overview.http[i]}
              />
            ))
          ) : (
            <Text size={4}>No related objects to show.</Text>
          )}
        </SectionContainer>

        <SectionContainer
          title="TCP"
          childrenContainerCss={{ gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))' }}
        >
          {overview?.tcp && hasResources.tcp ? (
            RESOURCES.map((i) => (
              <TraefikResourceStatsCard
                key={`tcp-${i}`}
                title={capitalizeFirstLetter(i)}
                data-testid={`section-tcp-${i}`}
                linkTo={`/tcp/${i}`}
                {...overview.tcp[i]}
              />
            ))
          ) : (
            <Text size={4}>No related objects to show.</Text>
          )}
        </SectionContainer>

        <SectionContainer
          title="UDP"
          childrenContainerCss={{ gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))' }}
        >
          {overview?.udp && hasResources.udp ? (
            RESOURCES.map((i) => (
              <TraefikResourceStatsCard
                key={`udp-${i}`}
                title={capitalizeFirstLetter(i)}
                data-testid={`section-udp-${i}`}
                linkTo={`/udp/${i}`}
                {...overview.udp[i]}
              />
            ))
          ) : (
            <Text size={4}>No related objects to show.</Text>
          )}
        </SectionContainer>

        <SectionContainer title="Features">
          {features.length
            ? features.map((i, idx) => {
                return <FeatureCard key={`feature-${idx}`} feature={i} />
              })
            : null}
        </SectionContainer>

        <SectionContainer title="Providers">
          {overview?.providers?.length ? (
            overview.providers.map((p, idx) => (
              <Card key={`provider-${idx}`} css={{ height: 125 }}>
                <Flex direction="column" align="center" gap={3} justify="center" css={{ height: '100%' }}>
                  <ProviderIcon name={p} size={52} />
                  <Text css={{ fontSize: '$4', fontWeight: 500, textAlign: 'center' }}>{p}</Text>
                </Flex>
              </Card>
            ))
          ) : (
            <Text size={4}>No related objects to show.</Text>
          )}
        </SectionContainer>
      </Flex>
    </Page>
  )
}

export const DashboardSkeleton = () => {
  return (
    <Page>
      <Flex direction="column" gap={6}>
        <SectionContainer title="Entrypoints" css={{ mt: 0 }}>
          {[...Array(5)].map((_, i) => (
            <FeatureCardSkeleton key={`entry-skeleton-${i}`} />
          ))}
        </SectionContainer>

        <SectionContainer
          title="HTTP"
          childrenContainerCss={{ gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))' }}
        >
          {[...Array(3)].map((_, i) => (
            <StatsCardSkeleton key={`http-skeleton-${i}`} />
          ))}
        </SectionContainer>

        <SectionContainer
          title="TCP"
          childrenContainerCss={{ gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))' }}
        >
          {[...Array(3)].map((_, i) => (
            <StatsCardSkeleton key={`tcp-skeleton-${i}`} />
          ))}
        </SectionContainer>

        <SectionContainer
          title="UDP"
          childrenContainerCss={{ gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))' }}
        >
          {[...Array(3)].map((_, i) => (
            <StatsCardSkeleton key={`udp-skeleton-${i}`} />
          ))}
        </SectionContainer>

        <SectionContainer title="Features">
          {[...Array(3)].map((_, i) => (
            <FeatureCardSkeleton key={`feature-skeleton-${i}`} />
          ))}
        </SectionContainer>

        <SectionContainer title="Providers">
          {[...Array(3)].map((_, i) => (
            <FeatureCardSkeleton key={`provider-skeleton-${i}`} />
          ))}
        </SectionContainer>
      </Flex>
    </Page>
  )
}
