import { Box, Card, Flex, Grid, Skeleton as FaencySkeleton, Text } from '@traefiklabs/faency'

import ResourceCard from 'components/resources/ResourceCard'

const FeatureCard = ({ feature }) => {
  const value = feature.value
  return (
    <ResourceCard title={feature.name}>
      <Box
        css={{
          px: '$3',
          borderRadius: '$2',
          py: '$2',
          backgroundColor: !value ? '$red6' : typeof value === 'boolean' ? '$green6' : '$gray6',
        }}
      >
        <Text
          css={{
            fontSize: '$10',
            fontWeight: 500,
            color: !value ? '$red10' : typeof value === 'boolean' ? '$green10' : '$gray10',
            textAlign: 'center',
          }}
        >
          {!value ? 'OFF' : typeof value === 'boolean' ? 'ON' : value}
        </Text>
      </Box>
    </ResourceCard>
  )
}

export const FeatureCardSkeleton = () => {
  return (
    <Grid gap={6} css={{ gridTemplateColumns: 'repeat(auto-fill, minmax(215px, 1fr))' }}>
      <Card css={{ minHeight: '125px' }}>
        <Flex justify="space-between" align="center" direction="column" css={{ height: '100%', p: '$2' }}>
          <FaencySkeleton css={{ width: 150, height: 13, mb: '$3' }} />
          <FaencySkeleton css={{ width: 80, height: 40 }} />
        </Flex>
      </Card>
    </Grid>
  )
}

export default FeatureCard
