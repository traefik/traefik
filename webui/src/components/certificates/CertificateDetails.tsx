import { Badge, Box, Flex, Link } from '@traefiklabs/faency'
import { useMemo } from 'react'

import CertExpiryBadge, { getCertExpiryStatus } from 'components/certificates/CertExpiryBadge'
import DetailsCard, { ValText } from 'components/resources/DetailsCard'

export const CertificateDetails = ({ certificate }: { certificate: Certificate.Info }) => {
  const validFrom = new Date(certificate.notBefore)
  const validUntil = new Date(certificate.notAfter)
  const certStatus = useMemo(() => getCertExpiryStatus(certificate.daysLeft), [certificate.daysLeft])

  const issuedToItems = [
    {
      key: 'Common Name',
      val: (
        <Link variant="blue" href={`//${certificate.commonName}`} target="_blank" rel="noopener noreferrer" css={{ fontSize: 'inherit' }}>
          {certificate.commonName}
        </Link>
      ),
    },
    {
      key: 'Status',
      val: (
        <Badge size="large" variant={certStatus.variant}>
          {certStatus.label}
        </Badge>
      ),
    },
    {
      key: 'Subject Alternative Names',
      val: (
        <Box>
          {certificate.sans.map((san, idx) => (
            <Box key={idx}>
              {san.startsWith('*.') ? (
                <ValText>{san}</ValText>
              ) : (
                <Link variant="blue" href={`//${san}`} target="_blank" rel="noopener noreferrer">
                  {san}
                </Link>
              )}
            </Box>
          ))}
        </Box>
      ),
    },
    { key: 'Organization', val: certificate.organization || '-' },
    { key: 'Country', val: certificate.country || '-' },
  ]

  const issuedByItems = [
    { key: 'Common Name', val: certificate.issuerCN || '-' },
    { key: 'Organization', val: certificate.issuerOrg || '-' },
    { key: 'Country', val: certificate.issuerCountry || '-' },
  ]

  const validityItems = [
    { key: 'Valid From', val: validFrom.toLocaleString() },
    { key: 'Valid Until', val: validUntil.toLocaleString() },
    {
      key: 'Expiry',
      val: <CertExpiryBadge daysLeft={certificate.daysLeft} />,
    },
  ]

  const technicalItems = [
    certificate.version && { key: 'Version', val: certificate.version },
    { key: 'Serial Number', val: certificate.serialNumber || 'N/A' },
    { key: 'Key Type', val: certificate.keyType || 'Unknown' },
    { key: 'Key Size', val: `${certificate.keySize || 0} bits` },
    { key: 'Signature Algorithm', val: certificate.signatureAlgorithm || 'Unknown' },
  ].filter(Boolean) as { key: string; val: string | React.ReactElement }[]

  const fingerprintItems = [
    { key: 'Certificate', val: certificate.certFingerprint || 'N/A' },
    { key: 'Public Key', val: certificate.publicKeyFingerprint || 'N/A' },
  ]

  return (
    <Flex direction="column" gap={2}>
      <DetailsCard title="Issued To" items={issuedToItems} keyColumns={1} minKeyWidth="200px" />
      <DetailsCard title="Issued By" items={issuedByItems} keyColumns={1} minKeyWidth="200px" />
      <DetailsCard title="Validity" items={validityItems} keyColumns={1} minKeyWidth="200px" />
      <DetailsCard title="Technical Details" items={technicalItems} keyColumns={1} minKeyWidth="200px" />
      <DetailsCard title="SHA-256 Fingerprints" items={fingerprintItems} keyColumns={1} minKeyWidth="200px" />
    </Flex>
  )
}

export default CertificateDetails
