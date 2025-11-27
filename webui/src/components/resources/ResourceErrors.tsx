import { Card, Flex, Skeleton } from '@traefiklabs/faency'

import { SectionTitle } from './DetailsCard'
import GenericTable from './GenericTable'

const ResourceErrors = ({ errors }: { errors: string[] }) => {
  return (
    <Flex direction="column" gap={2}>
      <SectionTitle title="Errors" />
      <Card>
        <GenericTable items={errors} status="error" copyable />
      </Card>
    </Flex>
  )
}

export const ResourceErrorsSkeleton = () => {
  return (
    <Flex direction="column" gap={2}>
      <Skeleton css={{ width: 200 }} />
      <Card css={{ width: '100%', height: 150, gap: '$3', display: 'flex', flexDirection: 'column' }}>
        {[...Array(4)].map((_, idx) => (
          <Skeleton key={`1-${idx}`} />
        ))}
      </Card>
    </Flex>
  )
}

export default ResourceErrors
