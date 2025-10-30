import { describe, it, expect, vi, beforeEach } from 'vitest'

import verifySignature from './scriptVerification'

describe('Script Signature Verification - Integration Tests', () => {
  let fetchMock: ReturnType<typeof vi.fn>

  const SCRIPT_URL = 'https://example.com/script.js'
  const SIGNATURE_URL = 'https://example.com/script.js.sig'
  const TEST_PUBLIC_KEY = 'MCowBQYDK2VwAyEAWH71OHphISjNK3mizCR/BawiDxc6IXT1vFHpBcxSIA0='
  const VALID_SCRIPT = "console.log('Hello from verified script!');"
  const VALID_SIGNATURE_HEX =
    '04c90fcd35caaf3cf4582a2767345f8cd9f6519e1ce79ebaeedbe0d5f671d762d1aa8ec258831557e2de0e47f224883f84eb5a0f22ec18eb7b8c48de3096d000'
  const CORRUPTED_SCRIPT = "console.log('Malicious code injected!');"

  beforeEach(() => {
    vi.clearAllMocks()
    fetchMock = vi.fn()
    globalThis.fetch = fetchMock
  })

  it('should verify a valid script with correct signature through real worker', async () => {
    fetchMock.mockImplementation((url: string) => {
      if (url === SCRIPT_URL) {
        return Promise.resolve(
          new Response(VALID_SCRIPT, {
            status: 200,
            headers: { 'Content-Type': 'application/javascript' },
          }),
        )
      }
      if (url === SIGNATURE_URL) {
        return Promise.resolve(
          new Response(VALID_SIGNATURE_HEX, {
            status: 200,
            headers: { 'Content-Type': 'text/plain' },
          }),
        )
      }
      return Promise.reject(new Error('Unexpected URL'))
    })

    const result = await verifySignature(SCRIPT_URL, SIGNATURE_URL, TEST_PUBLIC_KEY)

    expect(fetchMock).toHaveBeenCalledWith(SCRIPT_URL)
    expect(fetchMock).toHaveBeenCalledWith(SIGNATURE_URL)
    expect(result.verified).toBe(true)
    expect(result.scriptContent).toBeDefined()
  }, 15000)

  it('should reject a corrupted script with mismatched signature', async () => {
    fetchMock.mockImplementation((url: string) => {
      if (url === SCRIPT_URL) {
        return Promise.resolve(
          new Response(CORRUPTED_SCRIPT, {
            status: 200,
            headers: { 'Content-Type': 'application/javascript' },
          }),
        )
      }
      if (url === SIGNATURE_URL) {
        return Promise.resolve(
          new Response(VALID_SIGNATURE_HEX, {
            status: 200,
            headers: { 'Content-Type': 'text/plain' },
          }),
        )
      }
      return Promise.reject(new Error('Unexpected URL'))
    })

    const result = await verifySignature(SCRIPT_URL, SIGNATURE_URL, TEST_PUBLIC_KEY)

    expect(fetchMock).toHaveBeenCalledWith(SCRIPT_URL)
    expect(fetchMock).toHaveBeenCalledWith(SIGNATURE_URL)
    expect(result.verified).toBe(false)
    expect(result.scriptContent).toBeUndefined()
  }, 15000)

  it('should reject script with invalid signature format', async () => {
    fetchMock.mockImplementation((url: string) => {
      if (url === SCRIPT_URL) {
        return Promise.resolve(
          new Response(VALID_SCRIPT, {
            status: 200,
            headers: { 'Content-Type': 'application/javascript' },
          }),
        )
      }
      if (url === SIGNATURE_URL) {
        return Promise.resolve(
          new Response('not-a-valid-signature', {
            status: 200,
            headers: { 'Content-Type': 'text/plain' },
          }),
        )
      }
      return Promise.reject(new Error('Unexpected URL'))
    })

    const result = await verifySignature(SCRIPT_URL, SIGNATURE_URL, TEST_PUBLIC_KEY)

    expect(result.verified).toBe(false)
    expect(result.scriptContent).toBeUndefined()
  }, 15000)

  it('should reject script with wrong public key', async () => {
    const WRONG_PUBLIC_KEY = 'MCowBQYDK2VwAyEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=='

    fetchMock.mockImplementation((url: string) => {
      if (url === SCRIPT_URL) {
        return Promise.resolve(
          new Response(VALID_SCRIPT, {
            status: 200,
            headers: { 'Content-Type': 'application/javascript' },
          }),
        )
      }
      if (url === SIGNATURE_URL) {
        return Promise.resolve(
          new Response(VALID_SIGNATURE_HEX, {
            status: 200,
            headers: { 'Content-Type': 'text/plain' },
          }),
        )
      }
      return Promise.reject(new Error('Unexpected URL'))
    })

    const result = await verifySignature(SCRIPT_URL, SIGNATURE_URL, WRONG_PUBLIC_KEY)

    expect(result.verified).toBe(false)
    expect(result.scriptContent).toBeUndefined()
  }, 15000)

  it('should handle network failures when fetching script', async () => {
    fetchMock.mockImplementation(() =>
      Promise.resolve(
        new Response(null, {
          status: 404,
          statusText: 'Not Found',
        }),
      ),
    )

    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const result = await verifySignature(SCRIPT_URL, SIGNATURE_URL, TEST_PUBLIC_KEY)

    expect(result.verified).toBe(false)
    expect(result.scriptContent).toBeUndefined()
    expect(consoleErrorSpy).toHaveBeenCalled()

    consoleErrorSpy.mockRestore()
  }, 15000)
})
