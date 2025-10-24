import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

import { verifyScriptSignature } from './scriptVerification'

class MockWorker {
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((error: ErrorEvent) => void) | null = null
  postMessage = vi.fn()
  terminate = vi.fn()

  simulateMessage(data: unknown) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data }))
    }
  }

  simulateError(error: Error) {
    if (this.onerror) {
      this.onerror(new ErrorEvent('error', { error, message: error.message }))
    }
  }
}

describe('verifyScriptSignature', () => {
  let mockWorkerInstance: MockWorker

  beforeEach(() => {
    vi.clearAllMocks()

    mockWorkerInstance = new MockWorker()
    global.Worker = vi.fn(() => mockWorkerInstance) as unknown as typeof Worker
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should return true when verification succeeds', async () => {
    const publicKey = 'test-public-key'
    const scriptPath = 'https://example.com/script.js'
    const signaturePath = 'https://example.com/script.js.sig'

    const promise = verifyScriptSignature(publicKey, scriptPath, signaturePath)

    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(global.Worker).toHaveBeenCalled()

    expect(mockWorkerInstance.postMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        scriptUrl: scriptPath,
        signatureUrl: signaturePath,
        publicKey,
        requestId: expect.any(String),
      }),
    )

    const mockScriptContent = new ArrayBuffer(100)
    mockWorkerInstance.simulateMessage({
      success: true,
      verified: true,
      error: null,
      scriptContent: mockScriptContent,
    })

    const result = await promise

    expect(result).toEqual({ verified: true, scriptContent: mockScriptContent })
    expect(mockWorkerInstance.terminate).toHaveBeenCalled()
  })

  it('should return false when verification fails', async () => {
    const publicKey = 'test-public-key'
    const scriptPath = 'https://example.com/script.js'
    const signaturePath = 'https://example.com/script.js.sig'

    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const promise = verifyScriptSignature(publicKey, scriptPath, signaturePath)

    await new Promise((resolve) => setTimeout(resolve, 0))

    mockWorkerInstance.simulateMessage({
      success: false,
      verified: false,
      error: 'Signature verification failed',
    })

    const result = await promise

    expect(result).toEqual({ verified: false })
    expect(mockWorkerInstance.terminate).toHaveBeenCalled()
    expect(consoleErrorSpy).toHaveBeenCalledWith('Worker verification failed:', 'Signature verification failed')

    consoleErrorSpy.mockRestore()
  })

  it('should return false when worker throws an error', async () => {
    const publicKey = 'test-public-key'
    const scriptPath = 'https://example.com/script.js'
    const signaturePath = 'https://example.com/script.js.sig'

    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const promise = verifyScriptSignature(publicKey, scriptPath, signaturePath)

    await new Promise((resolve) => setTimeout(resolve, 0))

    // Simulate worker onerror event
    const error = new Error('Worker crashed')
    mockWorkerInstance.simulateError(error)

    const result = await promise

    expect(result).toEqual({ verified: false })
    expect(mockWorkerInstance.terminate).toHaveBeenCalled()
    expect(consoleErrorSpy).toHaveBeenCalledWith('Worker error:', expect.any(ErrorEvent))

    consoleErrorSpy.mockRestore()
  })
})
