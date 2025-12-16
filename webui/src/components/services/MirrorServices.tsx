import { Flex, Text } from '@traefiklabs/faency'
import { FiGlobe } from 'react-icons/fi'

import { getProviderFromName } from './Servers'

import ProviderIcon from 'components/icons/providers'
import { SectionTitle } from 'components/resources/DetailsCard'
import PaginatedTable from 'components/tables/PaginatedTable'

type MirrorServicesProps = {
  mirrors: Service.Mirror[]
  defaultProvider: string
}

const MirrorServices = ({ mirrors, defaultProvider }: MirrorServicesProps) => {
  return (
    <Flex direction="column" gap={2}>
      <SectionTitle icon={<FiGlobe size={20} />} title="Mirror Services" />
      <PaginatedTable
        data={mirrors.map((mirror) => ({
          name: mirror.name,
          percent: mirror.percent,
          provider: getProviderFromName(mirror.name, defaultProvider),
        }))}
        columns={[
          { key: 'name', header: 'Name' },
          { key: 'percent', header: 'Percent' },
          { key: 'provider', header: 'Provider' },
        ]}
        testId="mirror-services"
        renderCell={(key, value) => {
          if (key === 'provider') {
            return <ProviderIcon name={value as string} />
          }
          return <Text>{value}</Text>
        }}
      />
    </Flex>
  )
}

export default MirrorServices
