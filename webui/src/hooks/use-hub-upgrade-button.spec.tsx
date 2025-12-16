import { renderHook, waitFor } from '@testing-library/react'
import { ReactNode } from 'react'

import useHubUpgradeButton from './use-hub-upgrade-button'

import { VersionContext } from 'contexts/version'
import verifySignature from 'utils/workers/scriptVerification'

vi.mock('utils/workers/scriptVerification')

const mockVerifySignature = vi.mocked(verifySignature)

const createWrapper = (showHubButton: boolean) => {
  return ({ children }: { children: ReactNode }) => (
    <VersionContext.Provider value={{ showHubButton, version: '1.0.0' }}>{children}</VersionContext.Provider>
  )
}

describe('useHubUpgradeButton Hook', () => {
  let originalCreateObjectURL: typeof URL.createObjectURL
  let originalRevokeObjectURL: typeof URL.revokeObjectURL
  const mockBlobUrl = 'blob:http://localhost:3000/mock-blob-url'

  beforeEach(() => {
    originalCreateObjectURL = URL.createObjectURL
    originalRevokeObjectURL = URL.revokeObjectURL
    URL.createObjectURL = vi.fn(() => mockBlobUrl)
    URL.revokeObjectURL = vi.fn()
  })

  afterEach(() => {
    URL.createObjectURL = originalCreateObjectURL
    URL.revokeObjectURL = originalRevokeObjectURL
    vi.clearAllMocks()
  })

  it('should not verify script when showHubButton is false', async () => {
    renderHook(() => useHubUpgradeButton(), {
      wrapper: createWrapper(false),
    })

    await waitFor(() => {
      expect(mockVerifySignature).not.toHaveBeenCalled()
    })
  })

  it('should verify script and create blob URL when showHubButton is true and verification succeeds', async () => {
    const mockScriptContent = new ArrayBuffer(8)
    mockVerifySignature.mockResolvedValue({
      verified: true,
      scriptContent: mockScriptContent,
    })

    const { result } = renderHook(() => useHubUpgradeButton(), {
      wrapper: createWrapper(true),
    })

    await waitFor(() => {
      expect(mockVerifySignature).toHaveBeenCalledWith(
        'https://traefik.github.io/traefiklabs-hub-button-app/main-v1.js',
        'https://traefik.github.io/traefiklabs-hub-button-app/main-v1.js.sig',
        'MCowBQYDK2VwAyEAY0OZFFE5kSuqYK6/UprTL5RmvQ+8dpPTGMCw1MiO/Gs=',
      )
    })

    await waitFor(() => {
      expect(result.current.signatureVerified).toBe(true)
    })

    expect(result.current.scriptBlobUrl).toBe(mockBlobUrl)
    expect(URL.createObjectURL).toHaveBeenCalledWith(expect.any(Blob))
  })

  it('should set signatureVerified to false when verification fails', async () => {
    mockVerifySignature.mockResolvedValue({
      verified: false,
    })

    const { result } = renderHook(() => useHubUpgradeButton(), {
      wrapper: createWrapper(true),
    })

    await waitFor(() => {
      expect(mockVerifySignature).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(result.current.signatureVerified).toBe(false)
    })

    expect(result.current.scriptBlobUrl).toBeNull()
    expect(URL.createObjectURL).not.toHaveBeenCalled()
  })

  it('should handle verification errors gracefully', async () => {
    mockVerifySignature.mockRejectedValue(new Error('Verification failed'))

    const { result } = renderHook(() => useHubUpgradeButton(), {
      wrapper: createWrapper(true),
    })

    await waitFor(() => {
      expect(mockVerifySignature).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(result.current.signatureVerified).toBe(false)
    })

    expect(result.current.scriptBlobUrl).toBeNull()
  })

  it('should create blob with correct MIME type', async () => {
    const mockScriptContent = new ArrayBuffer(8)
    mockVerifySignature.mockResolvedValue({
      verified: true,
      scriptContent: mockScriptContent,
    })

    renderHook(() => useHubUpgradeButton(), {
      wrapper: createWrapper(true),
    })

    await waitFor(() => {
      expect(URL.createObjectURL).toHaveBeenCalled()
    })
    const blobCall = vi.mocked(URL.createObjectURL).mock.calls[0][0] as Blob
    expect(blobCall).toBeInstanceOf(Blob)
    expect(blobCall.type).toBe('application/javascript')
  })
})
