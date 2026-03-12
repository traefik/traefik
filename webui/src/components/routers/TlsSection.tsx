import { Badge, Box, Card, Flex, TabsContainer, TabsContent, TabsList, TabsTrigger, Text, Tooltip } from '@traefiklabs/faency'
import { useMemo } from 'react'

import { buildCertKey, useCertificate } from '../../hooks/use-certificates'
import { CertificateDetails } from '../certificates/CertificateDetails'

import TlsIcon from './TlsIcon'

import { EmptyIcon } from 'components/icons/EmptyIcon'
import { BooleanState, EmptyPlaceholder } from 'components/resources/DetailItemComponents'
import DetailsCard, { SectionTitle } from 'components/resources/DetailsCard'

type Props = {
  data?: Router.TLS
  rule?: string
}

// Extract domains from router rule (e.g., Host(`example.com`) || Host(`www.example.com`))
const extractDomainsFromRule = (rule: string): string[] =>
  [...rule.matchAll(/Host(?:SNI|Regexp)?\(`([^`]+)`\)/g)]
    .map(([, domain]) => domain)
    .filter((domain) => !/[{*[]/.test(domain))

const TlsSection = ({ data, rule }: Props) => {
  // Build display domains from explicit config or extract from rule if using certResolver
  const displayDomains = useMemo(() => {
    // 1. If explicit domains are configured, use those
    if (data?.domains && data.domains.length > 0) {
      return data.domains
    }

    // 2. If certResolver is set but no explicit domains, extract from rule
    if (data?.certResolver && rule) {
      const extracted = extractDomainsFromRule(rule)
      return extracted.map(domain => ({ main: domain, sans: [] as string[] }))
    }

    return []
  }, [data?.certResolver, data?.domains, rule])

  const items = useMemo(() => {
    if (data) {
      return [
        data?.options && { key: 'Options', val: data.options },
        { key: 'Passthrough', val: <BooleanState enabled={!!data.passthrough} /> },
        data?.certResolver && { key: 'Certificate resolver', val: data.certResolver },
      ].filter(Boolean) as { key: string; val: string | React.ReactElement }[]
    }
  }, [data])

  return (
    <Flex direction="column" gap={2}>
      <SectionTitle icon={<TlsIcon />} title="TLS" />
      {items?.length || displayDomains?.length ? (
        <>
          {items && items.length > 0 && <DetailsCard items={items} />}

          {displayDomains && displayDomains.length > 0 && (
            <Card css={{ p: '$4' }}>
              {displayDomains.length === 1 ? (
                <DomainCertificate domain={displayDomains[0]} />
              ) : (
                <TabsContainer defaultValue={displayDomains[0].main}>
                  <TabsList>
                    {displayDomains.map((domain) => {
                      const sansCount = domain.sans?.length || 0
                      return (
                        <TabsTrigger key={domain.main} value={domain.main}>
                          <Flex align="center" gap={2}>
                            <Text>{domain.main}</Text>
                            {sansCount > 0 && (
                              <Tooltip content={
                                <Flex direction="column" gap={1}>
                                  <Text css={{ color: 'currentColor', fontWeight: 600 }}>Subject Alternative Names:</Text>
                                  {domain.sans?.map((san) => (
                                    <Text key={san} css={{ color: 'currentColor', fontSize: '$1' }}>{san}</Text>
                                  ))}
                                </Flex>
                              }>
                                <Badge variant="gray">{sansCount}</Badge>
                              </Tooltip>
                            )}
                          </Flex>
                        </TabsTrigger>
                      )
                    })}
                  </TabsList>
                  {displayDomains.map((domain) => (
                    <TabsContent key={domain.main} value={domain.main} css={{ pt: '$4' }}>
                      <DomainCertificate domain={domain} />
                    </TabsContent>
                  ))}
                </TabsContainer>
              )}
            </Card>
          )}
        </>
      ) : (
        <Card>
          <Flex direction="column" align="center" justify="center" css={{ flexGrow: 1, textAlign: 'center', py: '$4' }}>
            <Box
              css={{
                width: 56,
                svg: {
                  width: '100%',
                  height: '100%',
                },
              }}
            >
              <EmptyIcon />
            </Box>
            <EmptyPlaceholder css={{ mt: '$3' }}>
              There is no
              <br />
              TLS configured
            </EmptyPlaceholder>
          </Flex>
        </Card>
      )}
    </Flex>
  )
}

// Helper component for displaying certificate for a single domain
const DomainCertificate = ({ domain }: { domain: Router.TlsDomain }) => {
  const certKey = useMemo(() => buildCertKey(domain.main, domain.sans), [domain])
  const { certificate, isLoading, error } = useCertificate(certKey)

  if (isLoading) {
    return <Text>Loading certificate...</Text>
  }

  if (error) {
    console.error('[TlsSection] Error fetching certificate:', error)
    return <Text variant="subtle">Error loading certificate for {domain.main}: {error.message || 'Unknown error'}</Text>
  }

  if (!certificate) {
    return <Text variant="subtle">No certificate found for {domain.main}</Text>
  }

  return <CertificateDetails certificate={certificate} />
}

export default TlsSection
