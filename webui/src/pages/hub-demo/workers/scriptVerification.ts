// Script verification using Web Worker in an isolated env
export async function verifyScriptSignature(
  publicKey: string,
  scriptPath: string,
  signaturePath: string,
): Promise<boolean> {
  return new Promise((resolve) => {
    const requestId = Math.random().toString(36).substring(2)
    const worker = new Worker(new URL('./scriptVerificationWorker.ts', import.meta.url), { type: 'module' })

    // Set timeout for worker
    const timeout = setTimeout(() => {
      worker.terminate()
      console.error('Script verification timeout')
      resolve(false)
    }, 30000)

    worker.onmessage = (event) => {
      clearTimeout(timeout)
      worker.terminate()

      const { success, verified, error } = event.data

      if (!success) {
        console.error('Worker verification failed:', error)
        resolve(false)
        return
      }

      resolve(verified === true)
    }

    worker.onerror = (error) => {
      clearTimeout(timeout)
      worker.terminate()
      console.error('Worker error:', error)
      resolve(false)
    }

    // Send verification request to worker
    worker.postMessage({
      requestId,
      scriptUrl: scriptPath,
      signatureUrl: signaturePath,
      publicKey,
    })
  })
}
