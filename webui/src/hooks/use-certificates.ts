import { useMemo } from 'react'
import useSWR from 'swr'

export const useCertificates = () => {
  const { data, error } = useSWR<Certificate.Raw[]>('/certificates')

  const certificates: Certificate.Info[] = useMemo(() => {
    if (!data) return []

    return data.map((cert) => ({
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

export const useCertificate = (certId: string) => {
  const { data, error } = useSWR<Certificate.Raw>(certId ? `/certificates/${certId}` : null)

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
    isLoading: !!certId && !error && !data,
  }
}
