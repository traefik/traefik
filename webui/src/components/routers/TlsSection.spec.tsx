import { screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'

import TlsSection from './TlsSection'

import * as useCertificates from '../../hooks/use-certificates'
import { CertificateInfo } from '../certificates/CertificateDetails'
import { renderWithProviders } from 'utils/test'

const mockCertificate: CertificateInfo = {
  commonName: 'example.com',
  sans: ['www.example.com'],
  notAfter: '2025-12-31T23:59:59Z',
  notBefore: '2024-01-01T00:00:00Z',
  serialNumber: '123456',
  daysLeft: 365,
  issuerOrg: 'Test CA',
  issuerCN: 'Test Root CA',
  status: 'enabled',
}

describe('<TlsSection />', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should display empty state when no TLS data provided', () => {
    renderWithProviders(<TlsSection />)

    // Text is split across <br/> so match with regex
    expect(screen.getByText(/There is no/)).toBeInTheDocument()
    expect(screen.getByText(/TLS configured/)).toBeInTheDocument()
  })

  it('should display TLS options, passthrough, and cert resolver', () => {
    const tlsData: Router.TLS = {
      options: 'default',
      passthrough: false,
      certResolver: 'letsencrypt',
      domains: [],
    }

    renderWithProviders(<TlsSection data={tlsData} />)

    const optionsLabels = screen.getAllByText('Options')
    expect(optionsLabels.length).toBeGreaterThan(0)
    expect(screen.getByText('default')).toBeInTheDocument()
    const passthroughLabels = screen.getAllByText('Passthrough')
    expect(passthroughLabels.length).toBeGreaterThan(0)
    const certResolverLabels = screen.getAllByText('Certificate resolver')
    expect(certResolverLabels.length).toBeGreaterThan(0)
    expect(screen.getByText('letsencrypt')).toBeInTheDocument()
  })

  it('should display certificate details for a single domain', async () => {
    vi.spyOn(useCertificates, 'useCertificate').mockReturnValue({
      certificate: mockCertificate,
      isLoading: false,
      error: null,
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [{ main: 'example.com', sans: ['www.example.com'] }],
    }

    renderWithProviders(<TlsSection data={tlsData} />)

    await waitFor(() => {
      expect(screen.getByText('Issued To')).toBeInTheDocument()
    })

    // Check that certificate details are rendered
    expect(screen.getByText('example.com')).toBeInTheDocument()
  })

  it('should display tabs for multiple domains', async () => {
    vi.spyOn(useCertificates, 'useCertificate').mockReturnValue({
      certificate: mockCertificate,
      isLoading: false,
      error: null,
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [
        { main: 'example.com', sans: ['www.example.com'] },
        { main: 'test.com', sans: [] },
      ],
    }

    renderWithProviders(<TlsSection data={tlsData} />)

    await waitFor(() => {
      // Domains appear in both tabs and certificate details
      expect(screen.getAllByText('example.com').length).toBeGreaterThan(0)
      expect(screen.getAllByText('test.com').length).toBeGreaterThan(0)
    })
  })

  it('should display SANs badge with tooltip', async () => {
    vi.spyOn(useCertificates, 'useCertificate').mockReturnValue({
      certificate: mockCertificate,
      isLoading: false,
      error: null,
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [
        { main: 'example.com', sans: ['www.example.com', 'api.example.com'] },
        { main: 'test.com', sans: [] },
      ],
    }

    renderWithProviders(<TlsSection data={tlsData} />)

    await waitFor(() => {
      // Should show badge with count of SANs
      expect(screen.getByText('2')).toBeInTheDocument()
    })
  })

  it('should extract domains from router rule when certResolver is set but no explicit domains', async () => {
    const spy = vi.spyOn(useCertificates, 'useCertificate')
    // Mock returns different certs based on domain in certKey
    spy.mockImplementation((certKey: string) => {
      const decoded = atob(certKey)
      if (decoded.includes('example.com')) {
        return {
          certificate: { ...mockCertificate, commonName: 'example.com' },
          isLoading: false,
          error: null,
        }
      }
      if (decoded.includes('www.example.com')) {
        return {
          certificate: { ...mockCertificate, commonName: 'www.example.com', sans: [] },
          isLoading: false,
          error: null,
        }
      }
      return { certificate: null, isLoading: false, error: null }
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [],
    }

    const rule = 'Host(`example.com`) || Host(`www.example.com`)'

    renderWithProviders(<TlsSection data={tlsData} rule={rule} />)

    await waitFor(() => {
      // Should extract and display domains from rule (appear in tabs and certificate details)
      expect(screen.getAllByText('example.com').length).toBeGreaterThan(0)
      expect(screen.getAllByText('www.example.com').length).toBeGreaterThan(0)
    })
  })

  it('should not extract domains with wildcards or regex patterns', async () => {
    vi.spyOn(useCertificates, 'useCertificate').mockReturnValue({
      certificate: null,
      isLoading: false,
      error: null,
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [],
    }

    const rule = 'HostRegexp(`*.example.com`) || Host(`{subdomain:[a-z]+}.example.com`)'

    renderWithProviders(<TlsSection data={tlsData} rule={rule} />)

    await waitFor(() => {
      // Should not extract domains with wildcards/regex
      expect(screen.queryByText('*.example.com')).not.toBeInTheDocument()
    })

    // Should show only the TLS config section (certResolver) but no certificate details since no domains extracted
    expect(screen.getByText('letsencrypt')).toBeInTheDocument()
  })

  it('should display loading state while fetching certificate', () => {
    vi.spyOn(useCertificates, 'useCertificate').mockReturnValue({
      certificate: null,
      isLoading: true,
      error: null,
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [{ main: 'example.com', sans: [] }],
    }

    renderWithProviders(<TlsSection data={tlsData} />)

    expect(screen.getByText('Loading certificate...')).toBeInTheDocument()
  })

  it('should display error state when certificate fetch fails', () => {
    const mockError = new Error('Network error')
    vi.spyOn(useCertificates, 'useCertificate').mockReturnValue({
      certificate: null,
      isLoading: false,
      error: mockError,
    })
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [{ main: 'example.com', sans: [] }],
    }

    renderWithProviders(<TlsSection data={tlsData} />)

    expect(screen.getByText(/Error loading certificate for example.com/)).toBeInTheDocument()
    expect(screen.getByText(/Network error/)).toBeInTheDocument()
    expect(consoleErrorSpy).toHaveBeenCalledWith('[TlsSection] Error fetching certificate:', mockError)

    consoleErrorSpy.mockRestore()
  })

  it('should display message when certificate not found', () => {
    vi.spyOn(useCertificates, 'useCertificate').mockReturnValue({
      certificate: null,
      isLoading: false,
      error: null,
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [{ main: 'example.com', sans: [] }],
    }

    renderWithProviders(<TlsSection data={tlsData} />)

    expect(screen.getByText('No certificate found for example.com')).toBeInTheDocument()
  })

  it('should display passthrough as enabled', () => {
    const tlsData: Router.TLS = {
      passthrough: true,
      domains: [],
    }

    renderWithProviders(<TlsSection data={tlsData} />)

    const passthroughLabels = screen.getAllByText('Passthrough')
    expect(passthroughLabels.length).toBeGreaterThan(0)
    // BooleanState renders an icon when enabled
    const passthroughValue = screen.getAllByText('Passthrough')[0].closest('div')?.nextElementSibling
    expect(passthroughValue).toBeTruthy()
  })

  it('should prefer explicit domains over rule extraction', async () => {
    const spy = vi.spyOn(useCertificates, 'useCertificate')
    spy.mockImplementation((certKey: string) => {
      const decoded = atob(certKey)
      if (decoded.includes('explicit.com')) {
        return {
          certificate: { ...mockCertificate, commonName: 'explicit.com', sans: [] },
          isLoading: false,
          error: null,
        }
      }
      return { certificate: null, isLoading: false, error: null }
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [{ main: 'explicit.com', sans: [] }],
    }

    const rule = 'Host(`from-rule.com`)'

    renderWithProviders(<TlsSection data={tlsData} rule={rule} />)

    await waitFor(() => {
      // Should use explicit domain, not extract from rule
      expect(screen.getByText('explicit.com')).toBeInTheDocument()
      expect(screen.queryByText('from-rule.com')).not.toBeInTheDocument()
    })
  })

  it('should handle HostSNI matcher for TCP routers', async () => {
    const spy = vi.spyOn(useCertificates, 'useCertificate')
    spy.mockImplementation((certKey: string) => {
      const decoded = atob(certKey)
      if (decoded.includes('sni-example.com')) {
        return {
          certificate: { ...mockCertificate, commonName: 'sni-example.com', sans: [] },
          isLoading: false,
          error: null,
        }
      }
      return { certificate: null, isLoading: false, error: null }
    })

    const tlsData: Router.TLS = {
      certResolver: 'letsencrypt',
      domains: [],
    }

    const rule = 'HostSNI(`sni-example.com`)'

    renderWithProviders(<TlsSection data={tlsData} rule={rule} />)

    await waitFor(() => {
      expect(screen.getByText('sni-example.com')).toBeInTheDocument()
    })
  })
})
