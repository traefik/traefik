import * as useCertificates from '../../hooks/use-certificates'
import { Certificate } from './Certificate'
import { renderWithProviders } from 'utils/test'

describe('<CertificatePage />', () => {
  it('should render the loading state initially', () => {
    vi.spyOn(useCertificates, 'useCertificate').mockImplementation(() => ({
      certificate: null,
      error: null,
      isLoading: true,
    }))

    const { getByText } = renderWithProviders(<Certificate />, {
      route: '/certificates/dW5rbm93bi1jZXJ0LWtleQ==',
      withPage: true,
    })
    
    expect(getByText('Loading certificate...')).toBeInTheDocument()
  })

  it('should render error message when API returns error', () => {
    vi.spyOn(useCertificates, 'useCertificate').mockImplementation(() => ({
      certificate: null,
      error: new Error('Internal Server Error'),
      isLoading: false,
    }))

    const { getByText } = renderWithProviders(<Certificate />, {
      route: '/certificates/c29tZS1jZXJ0',
      withPage: true,
    })

    expect(getByText('Error loading certificate: Internal Server Error')).toBeInTheDocument()
  })

  it('should render "Certificate not found" when certificate is null', () => {
    vi.spyOn(useCertificates, 'useCertificate').mockImplementation(() => ({
      certificate: null,
      error: null,
      isLoading: false,
    }))

    const { getByText } = renderWithProviders(<Certificate />, {
      route: '/certificates/bm90Zm91bmQ=',
      withPage: true,
    })

    expect(getByText('Certificate not found')).toBeInTheDocument()
  })

  it('should render certificate details successfully', () => {
    const mockCertificate = {
      name: 'dGVzdC1jZXJ0',
      commonName: 'test.com',
      sans: ['test.com', 'www.test.com'],
      issuerOrg: 'Test CA',
      issuerCN: 'Test Root CA',
      notAfter: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString(),
      notBefore: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
      status: 'enabled' as const,
      serialNumber: '123456',
    }

    vi.spyOn(useCertificates, 'useCertificate').mockImplementation(() => ({
      certificate: { ...mockCertificate, daysLeft: 365 },
      error: null,
      isLoading: false,
    }))

    const { getByText } = renderWithProviders(<Certificate />, {
      route: '/certificates/dGVzdC1jZXJ0',
      withPage: true,
    })

    // Check for actual rendered content
    expect(getByText('Certificate')).toBeInTheDocument()
    expect(getByText('Issued To')).toBeInTheDocument()
    expect(getByText('Issued By')).toBeInTheDocument()
    expect(getByText('www.test.com')).toBeInTheDocument()
    expect(getByText('Test CA')).toBeInTheDocument()
    expect(getByText('Test Root CA')).toBeInTheDocument()
  })
})
