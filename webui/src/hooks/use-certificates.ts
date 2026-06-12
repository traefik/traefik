import { useMemo } from 'react'
import useSWR from 'swr'

const msPerDay = 1000 * 60 * 60 * 24;

export const computeDaysLeft = (notAfter: string): number =>
  Math.floor((new Date(notAfter).getTime() - Date.now()) / msPerDay)

const getCertificateInfo = (cert: Certificate.Raw): Certificate.Info => ({
  ...cert,
  daysLeft: computeDaysLeft(cert.notAfter),
})

export const useCertificates = () => {
  const { data, error } = useSWR<Certificate.Raw[]>('/certificates')

  const certificates = useMemo<Certificate.Info[]>(
    () => data?.map(getCertificateInfo) ?? [],
    [data],
  )

  return {
    certificates,
    error,
    isLoading: !error && !data,
  }
}

export const useCertificate = (certId: string) => {
  const { data, error } = useSWR<Certificate.Raw>(certId ? `/certificates/${certId}` : null)

  const certificate = useMemo<Certificate.Info | null>(
    () => data
      ? getCertificateInfo(data)
      : null,
    [data],
  )

  return {
    certificate,
    error,
    isLoading: !!certId && !error && !data,
  }
}
