import { Card, H2 } from '@traefiklabs/faency'
import { useParams } from 'react-router-dom'

import { CertificateDetails } from '../../components/certificates/CertificateDetails'
import { useCertificate } from '../../hooks/use-certificates'

import PageTitle from 'layout/PageTitle'

export const Certificate = () => {
  const { name } = useParams<{ name: string }>()
  const { certificate, isLoading, error } = useCertificate(name || '')

  if (isLoading) {
    return <div>Loading certificate...</div>
  }

  if (error) {
    return <div>Error loading certificate: {error.message}</div>
  }

  if (!certificate) {
    return <div>Certificate not found</div>
  }

  return (
    <div>
      <PageTitle title={`Certificate: ${certificate.commonName}`} />
      
      <Card css={{ p: '$4' }}>
        <H2 css={{ mb: '$4' }}>Certificate</H2>
        <CertificateDetails certificate={certificate} />
      </Card>
    </div>
  )
}
