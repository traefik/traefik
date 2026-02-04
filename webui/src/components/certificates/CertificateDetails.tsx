import { Badge, Box, Flex, Grid, Text } from '@traefiklabs/faency'

// Matches the API response structure exactly
export interface CertificateInfo {
  commonName: string
  sans: string[]
  issuerOrg?: string
  issuerCN?: string
  issuerCountry?: string
  organization?: string
  country?: string
  serialNumber?: string
  notBefore: string
  notAfter: string
  daysLeft: number
  version?: string
  keyType?: string
  keySize?: number
  signatureAlgorithm?: string
  certFingerprint?: string
  publicKeyFingerprint?: string
  status?: string
  resolver?: string
}

interface CertificateDetailsProps {
  certificate: CertificateInfo
}

const getCertStatusColor = (daysLeft: number) => {
  if (daysLeft < 0) return { variant: 'red' as const, label: 'EXPIRED' }
  if (daysLeft < 30) return { variant: 'orange' as const, label: 'Expiring Soon' }
  return { variant: 'green' as const, label: 'Valid' }
}

export const CertificateDetails = ({ certificate }: CertificateDetailsProps) => {
  const validFrom = new Date(certificate.notBefore)
  const validUntil = new Date(certificate.notAfter)

  return (
    <Flex direction="column" gap={4}>
      <Box>
        <Text css={{ fontWeight: 600, mb: '$3' }}>Issued To</Text>
        <Grid css={{ gridTemplateColumns: '200px 1fr', gap: '$3' }}>
          <Text variant="subtle">Common Name:</Text>
          <div>
            <a href={`//${certificate.commonName}`} target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none', display: 'inline' }}>
              <Text css={{ color: '$primary', fontWeight: 600, '&:hover': { textDecoration: 'underline' } }}>{certificate.commonName}</Text>
            </a>
          </div>
          
          <Text variant="subtle">Status:</Text>
          <div>
            <Badge size="small" variant={getCertStatusColor(certificate.daysLeft).variant}>
              {getCertStatusColor(certificate.daysLeft).label}
            </Badge>
          </div>
          
          <Text variant="subtle">Subject Alternative Names:</Text>
          <div>
            {certificate.sans.map((san, idx) => (
              <div key={idx}>
                {san.startsWith('*.') ? (
                  <Text>{san}</Text>
                ) : (
                  <a href={`//${san}`} target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none', display: 'inline-block' }}>
                    <Text css={{ color: '$primary', '&:hover': { textDecoration: 'underline' } }}>{san}</Text>
                  </a>
                )}
              </div>
            ))}
          </div>
          
          <Text variant="subtle">Organization:</Text>
          <Text>{certificate.organization || '-'}</Text>
          
          <Text variant="subtle">Country:</Text>
          <Text>{certificate.country || '-'}</Text>
        </Grid>
      </Box>

      <Box>
        <Text css={{ fontWeight: 600, mb: '$3' }}>Issued By</Text>
        <Grid css={{ gridTemplateColumns: '200px 1fr', gap: '$3' }}>
          <Text variant="subtle">Common Name:</Text>
          <Text>{certificate.issuerCN || '-'}</Text>
          
          <Text variant="subtle">Organization:</Text>
          <Text>{certificate.issuerOrg || '-'}</Text>
          
          <Text variant="subtle">Country:</Text>
          <Text>{certificate.issuerCountry || '-'}</Text>
        </Grid>
      </Box>

      <Box>
        <Text css={{ fontWeight: 600, mb: '$3' }}>Validity</Text>
        <Grid css={{ gridTemplateColumns: '200px 1fr', gap: '$3' }}>
          <Text variant="subtle">Valid From:</Text>
          <Text>{validFrom.toLocaleString()}</Text>
          
          <Text variant="subtle">Valid Until:</Text>
          <Text>{validUntil.toLocaleString()}</Text>
          
          <Text variant="subtle">Expiry:</Text>
          <div>
            <Badge size="small" variant={getCertStatusColor(certificate.daysLeft).variant}>
              {certificate.daysLeft < 0 ? 'EXPIRED' : `${certificate.daysLeft} days left`}
            </Badge>
          </div>
        </Grid>
      </Box>

      <Box>
        <Text css={{ fontWeight: 600, mb: '$3' }}>Technical Details</Text>
        <Grid css={{ gridTemplateColumns: '200px 1fr', gap: '$3' }}>
          {certificate.version && (
            <>
              <Text variant="subtle">Version:</Text>
              <Text>{certificate.version}</Text>
            </>
          )}
          
          <Text variant="subtle">Serial Number:</Text>
          <Text css={{ fontFamily: 'monospace', fontSize: '$2' }}>{certificate.serialNumber || 'N/A'}</Text>
          
          <Text variant="subtle">Key Type:</Text>
          <Text>{certificate.keyType || 'Unknown'}</Text>
          
          <Text variant="subtle">Key Size:</Text>
          <Text>{certificate.keySize || 0} bits</Text>
          
          <Text variant="subtle">Signature Algorithm:</Text>
          <Text>{certificate.signatureAlgorithm || 'Unknown'}</Text>
        </Grid>
      </Box>

      <Box>
        <Text css={{ fontWeight: 600, mb: '$3' }}>SHA-256 Fingerprints</Text>
        <Grid css={{ gridTemplateColumns: '200px 1fr', gap: '$3' }}>
          <Text variant="subtle">Certificate:</Text>
          <Text css={{ fontFamily: 'monospace', fontSize: '$2' }}>{certificate.certFingerprint || 'N/A'}</Text>
          
          <Text variant="subtle">Public Key:</Text>
          <Text css={{ fontFamily: 'monospace', fontSize: '$2' }}>{certificate.publicKeyFingerprint || 'N/A'}</Text>
        </Grid>
      </Box>
    </Flex>
  )
}

export default CertificateDetails
