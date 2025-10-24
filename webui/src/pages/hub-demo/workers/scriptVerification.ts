export interface VerificationResult {
  verified: boolean
  scriptContent?: ArrayBuffer
}

export async function verifyScriptSignature(
  publicKey: string,
  scriptPath: string,
  signaturePath: string,
): Promise<VerificationResult> {
  return new Promise((resolve) => {
    const requestId = Math.random().toString(36).substring(2)
    const worker = new Worker(new URL('./scriptVerificationWorker.ts', import.meta.url), { type: 'module' })

    // Set timeout for worker
    const timeout = setTimeout(() => {
      worker.terminate()
      console.error('Script verification timeout')
      resolve({ verified: false })
    }, 30000)

    worker.onmessage = (event) => {
      clearTimeout(timeout)
      worker.terminate()

      const { success, verified, error, scriptContent } = event.data

      if (!success) {
        console.error('Worker verification failed:', error)
        resolve({ verified: false })
        return
      }

      resolve({
        verified: verified === true,
        scriptContent: verified ? scriptContent : undefined,
      })
    }

    worker.onerror = (error) => {
      clearTimeout(timeout)
      worker.terminate()
      console.error('Worker error:', error)
      resolve({ verified: false })
    }

    worker.postMessage({
      requestId,
      scriptUrl: scriptPath,
      signatureUrl: signaturePath,
      publicKey,
    })
  })
}
