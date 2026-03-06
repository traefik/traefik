import { Box, Card, Flex, H1, Skeleton, Text } from '@traefiklabs/faency'
import { useParams } from 'react-router-dom'

import { CertificateDetails } from '../../components/certificates/CertificateDetails'
import { useCertificate } from '../../hooks/use-certificates'

import { DetailsCardSkeleton } from 'components/resources/DetailsCard'
import { ResourceStatus } from 'components/resources/ResourceStatus'
import PageTitle from 'layout/PageTitle'
import { NotFound } from 'pages/NotFound'

export const Certificate = () => {
  const { name } = useParams<{ name: string }>()
  const { certificate, isLoading, error } = useCertificate(name || '')

  if (isLoading) {
    return (
      <Box>
        <PageTitle title={name || ''} />
        <Skeleton css={{ height: '$7', width: '320px', mb: '$7' }} data-testid="skeleton" />
        <Flex direction="column" gap={6}>
          <DetailsCardSkeleton keyColumns={1} rows={5} />
        </Flex>
      </Box>
    )
  }

  if (error) {
    return (
      <>
        <PageTitle title={certificate?.commonName || name || ''} />
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Certificate right now. Please, try again later.
        </Text>
      </>
    )
  }

  if (!certificate) {
    return <NotFound />
  }

  return (
    <>
      <PageTitle title={`Certificate: ${certificate.commonName}`} />
      <Flex gap={2} align="center"  css={{ mb: '$4' }}>
        <H1>{certificate.commonName}</H1>
        <ResourceStatus status={certificate.status || 'disabled'} />
      </Flex>
      <CertificateDetails certificate={certificate} />
    </>
  )
}
