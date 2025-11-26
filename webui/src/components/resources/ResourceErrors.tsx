import { Card, Flex } from '@traefiklabs/faency'

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

export default ResourceErrors
