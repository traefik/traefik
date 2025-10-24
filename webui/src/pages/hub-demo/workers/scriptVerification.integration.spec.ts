import { describe, it, expect, vi, beforeEach } from 'vitest'

import verifySignature from './scriptVerification'

describe('Script Signature Verification - Integration Tests', () => {
  let fetchMock: ReturnType<typeof vi.fn>

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
    const scriptUrl = 'https://example.com/script.js'
    const signatureUrl = 'https://example.com/script.js.sig'

    fetchMock.mockImplementation((url: string) => {
      if (url === scriptUrl) {
        return Promise.resolve(
          new Response(VALID_SCRIPT, {
            status: 200,
            headers: { 'Content-Type': 'application/javascript' },
          }),
        )
      }
      if (url === signatureUrl) {
        return Promise.resolve(
          new Response(VALID_SIGNATURE_HEX, {
            status: 200,
            headers: { 'Content-Type': 'text/plain' },
          }),
        )
      }
      return Promise.reject(new Error('Unexpected URL'))
    })

    const result = await verifySignature(scriptUrl, signatureUrl, TEST_PUBLIC_KEY)

    expect(fetchMock).toHaveBeenCalledWith(scriptUrl)
    expect(fetchMock).toHaveBeenCalledWith(signatureUrl)
    expect(result).toBe(true)
  }, 15000)

  it('should reject a corrupted script with mismatched signature', async () => {
    const scriptUrl = 'https://example.com/script.js'
    const signatureUrl = 'https://example.com/script.js.sig'

    fetchMock.mockImplementation((url: string) => {
      if (url === scriptUrl) {
        return Promise.resolve(
          new Response(CORRUPTED_SCRIPT, {
            status: 200,
            headers: { 'Content-Type': 'application/javascript' },
          }),
        )
      }
      if (url === signatureUrl) {
        return Promise.resolve(
          new Response(VALID_SIGNATURE_HEX, {
            status: 200,
            headers: { 'Content-Type': 'text/plain' },
          }),
        )
      }
      return Promise.reject(new Error('Unexpected URL'))
    })

    const result = await verifySignature(scriptUrl, signatureUrl, TEST_PUBLIC_KEY)

    expect(fetchMock).toHaveBeenCalledWith(scriptUrl)
    expect(fetchMock).toHaveBeenCalledWith(signatureUrl)
    expect(result).toBe(false)
  }, 15000)

  it('should reject script with invalid signature format', async () => {
    const scriptUrl = 'https://example.com/script.js'
    const signatureUrl = 'https://example.com/script.js.sig'

    fetchMock.mockImplementation((url: string) => {
      if (url === scriptUrl) {
        return Promise.resolve(
          new Response(VALID_SCRIPT, {
            status: 200,
            headers: { 'Content-Type': 'application/javascript' },
          }),
        )
      }
      if (url === signatureUrl) {
        return Promise.resolve(
          new Response('not-a-valid-signature', {
            status: 200,
            headers: { 'Content-Type': 'text/plain' },
          }),
        )
      }
      return Promise.reject(new Error('Unexpected URL'))
    })

    const result = await verifySignature(scriptUrl, signatureUrl, TEST_PUBLIC_KEY)

    expect(result).toBe(false)
  }, 15000)

  it('should reject script with wrong public key', async () => {
    const scriptUrl = 'https://example.com/script.js'
    const signatureUrl = 'https://example.com/script.js.sig'

    const WRONG_PUBLIC_KEY = 'MCowBQYDK2VwAyEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=='

    fetchMock.mockImplementation((url: string) => {
      if (url === scriptUrl) {
        return Promise.resolve(
          new Response(VALID_SCRIPT, {
            status: 200,
            headers: { 'Content-Type': 'application/javascript' },
          }),
        )
      }
      if (url === signatureUrl) {
        return Promise.resolve(
          new Response(VALID_SIGNATURE_HEX, {
            status: 200,
            headers: { 'Content-Type': 'text/plain' },
          }),
        )
      }
      return Promise.reject(new Error('Unexpected URL'))
    })

    const result = await verifySignature(scriptUrl, signatureUrl, WRONG_PUBLIC_KEY)

    expect(result).toBe(false)
  }, 15000)

  it('should handle network failures when fetching script', async () => {
    const scriptUrl = 'https://example.com/script.js'
    const signatureUrl = 'https://example.com/script.js.sig'

    fetchMock.mockImplementation(() =>
      Promise.resolve(
        new Response(null, {
          status: 404,
          statusText: 'Not Found',
        }),
      ),
    )

    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const result = await verifySignature(scriptUrl, signatureUrl, TEST_PUBLIC_KEY)

    expect(result).toBe(false)
    expect(consoleErrorSpy).toHaveBeenCalled()

    consoleErrorSpy.mockRestore()
  }, 15000)
})
