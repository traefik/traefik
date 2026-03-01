import { Flex } from '@traefiklabs/faency'
import { FiGlobe } from 'react-icons/fi'

import { getProviderFromName } from './utils'

import ProviderIcon from 'components/icons/providers'
import { SectionTitle } from 'components/resources/DetailsCard'
import PaginatedTable from 'components/tables/PaginatedTable'

type WeightedServicesProps = {
  services: Service.WeightedService[]
  defaultProvider: string
}

const WeightedServices = ({ services, defaultProvider }: WeightedServicesProps) => {
  return (
    <Flex direction="column" gap={2}>
      <SectionTitle icon={<FiGlobe size={20} />} title="Services" />
      <PaginatedTable
        data={services.map((service) => ({
          name: service.name,
          weight: service.weight,
          provider: getProviderFromName(service.name, defaultProvider),
        }))}
        columns={[
          { key: 'name', header: 'Name' },
          { key: 'weight', header: 'Weight' },
          { key: 'provider', header: 'Provider' },
        ]}
        testId="weighted-services"
        renderCell={(key, value) => {
          if (key === 'provider') {
            return <ProviderIcon name={value as string} />
          }
          return value
        }}
      />
    </Flex>
  )
}

export default WeightedServices
