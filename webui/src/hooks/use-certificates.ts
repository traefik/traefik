import { useMemo } from 'react'
import useSWR from 'swr'

export const computeDaysLeft = (notAfter: string): number =>
  Math.floor((new Date(notAfter).getTime() - Date.now()) / (1000 * 60 * 60 * 24))

export const useCertificates = () => {
  const { data, error } = useSWR<Certificate.Raw[]>('/certificates')

  const certificates: Certificate.Info[] = useMemo(() => {
    if (!data) return []

    return data.map((cert) => ({
      ...cert,
      daysLeft: computeDaysLeft(cert.notAfter),
    }))
  }, [data])

  return {
    certificates,
    error,
    isLoading: !error && !data,
  }
}

export const useCertificate = (certId: string) => {
  const { data, error } = useSWR<Certificate.Raw>(certId ? `/certificates/${certId}` : null)

  const certificate: Certificate.Info | null = useMemo(() => {
    if (!data) return null

    return {
      ...data,
      daysLeft: computeDaysLeft(data.notAfter),
    }
  }, [data])

  return {
    certificate,
    error,
    isLoading: !!certId && !error && !data,
  }
}
