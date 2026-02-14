import { useMemo } from 'react'
import useSWR from 'swr'

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
  const { data, error } = useSWR<Certificate.Raw[]>('/certificates')

  const certificates: Certificate.Info[] = useMemo(() => {
    if (!data) return []

    return data.map(cert => ({
      ...cert,
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
  const { data, error } = useSWR<Certificate.Raw>(
    certKey ? `/certificates/${encodeURIComponent(certKey)}` : null
  )

  const certificate: Certificate.Info | null = useMemo(() => {
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
