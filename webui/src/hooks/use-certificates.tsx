import { useMemo } from 'react'
import useSWR from 'swr'

export interface Certificate {
  name?: string // certKey (sorted SANs joined with comma)
  sans: string[]
  notAfter: string
  notBefore: string
  serialNumber?: string
  commonName: string
  issuer?: string
  issuerOrg?: string
  issuerCN?: string
  issuerCountry?: string
  organization?: string
  country?: string
  subject?: string
  version?: string
  keyType?: string
  keySize?: number
  signatureAlgorithm?: string
  certFingerprint?: string
  publicKeyFingerprint?: string
  status?: 'enabled' | 'disabled' | 'warning'
  resolver?: string
  usedBy?: string[]
}

/**
 * Build a certificate key from domains (main + SANs)
 * Returns base64-encoded string of sorted, comma-separated domains
 * Deduplicates and lowercases domains to match backend behavior
 */
export const buildCertKey = (main: string, sans?: string[]): string => {
  const allDomains = [main, ...(sans || [])]
  // Deduplicate using Set (lowercased), then sort and join
  const uniqueDomains = Array.from(new Set(allDomains.map(d => d.toLowerCase()))).sort().join(',')
  return btoa(uniqueDomains)
}

export const useCertificates = () => {
  const { data, error } = useSWR<Certificate[]>('/certificates')

  const certificates = useMemo(() => {
    if (!data) return []
    
    return data.map(cert => ({
      ...cert,
      // Calculate days left from notAfter
      daysLeft: cert.notAfter
        ? Math.floor((new Date(cert.notAfter).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
        : 0,
    }))
  }, [data])

  return {
    certificates,
    error,
    isLoading: !error && !data,
  }
}

export const useCertificate = (certKey: string) => {
  const { data, error } = useSWR<Certificate>(
    certKey ? `/certificates/${encodeURIComponent(certKey)}` : null
  )

  const certificate = useMemo(() => {
    if (!data) return null

    return {
      ...data,
      daysLeft: data.notAfter
        ? Math.floor((new Date(data.notAfter).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
        : 0,
    }
  }, [data])

  return {
    certificate,
    error,
    isLoading: !error && !data,
  }
}
