import { Badge, Box, Flex, Text } from '@traefiklabs/faency'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

import GenericTable from 'components/resources/GenericTable'
import Page from 'layout/Page'
import useFetchWithPagination from 'hooks/use-fetch-with-pagination'
import { BASE_PATH } from 'libs/utils'

const TlsCertificates = () => {
  const [searchParams] = useSearchParams()

  const { data, isLoading, error } = useFetchWithPagination<API.Certificate>({
    url: `${BASE_PATH}/tls/certificates`,
    searchParams,
  })

  const columns = useMemo(
    () => [
      {
        id: 'name',
        header: 'Name',
        accessorKey: 'name',
        cell: ({ row }: { row: { original: API.Certificate } }) => (
          <Text css={{ fontWeight: '$semiBold' }}>{row.original.name}</Text>
        ),
      },
      {
        id: 'domains',
        header: 'Domains',
        accessorKey: 'domains',
        cell: ({ row }: { row: { original: API.Certificate } }) => (
          <Flex gap="1" wrap="wrap">
            {row.original.domains.map((domain, index) => (
              <Badge key={index} variant="secondary" css={{ fontSize: '$1' }}>
                {domain}
              </Badge>
            ))}
          </Flex>
        ),
      },
      {
        id: 'expiration',
        header: 'Expiration',
        accessorKey: 'expiration',
        cell: ({ row }: { row: { original: API.Certificate } }) => {
          const expirationDate = new Date(row.original.expiration)
          const now = new Date()
          const daysUntilExpiry = Math.ceil((expirationDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
          
          let variant: 'default' | 'green' | 'yellow' | 'red' = 'default'
          if (daysUntilExpiry < 0) {
            variant = 'red'
          } else if (daysUntilExpiry < 30) {
            variant = 'yellow'
          } else {
            variant = 'green'
          }

          return (
            <Box>
              <Text css={{ fontSize: '$2' }}>
                {expirationDate.toLocaleDateString()}
              </Text>
              <Badge variant={variant} css={{ ml: '$2', fontSize: '$1' }}>
                {daysUntilExpiry < 0 ? 'Expired' : `${daysUntilExpiry} days`}
              </Badge>
            </Box>
          )
        },
      },
    ],
    [],
  )

  return (
    <Page title="TLS Certificates">
      <GenericTable
        data={data}
        columns={columns}
        isLoading={isLoading}
        error={error}
        emptyMessage="No TLS certificates found"
      />
    </Page>
  )
}

export default TlsCertificates
