// Script verification worker
// Runs in isolated context for secure verification

import { verify } from '@noble/ed25519'
import * as ed25519 from '@noble/ed25519'
import { sha512 } from '@noble/hashes/sha2.js'

// Set up SHA-512 for @noble/ed25519 v3.x
ed25519.hashes.sha512 = sha512
ed25519.hashes.sha512Async = (m) => Promise.resolve(sha512(m))

function base64ToArrayBuffer(base64: string): ArrayBuffer {
  try {
    // @ts-expect-error - fromBase64 is not yet in all TypeScript lib definitions
    const bytes = Uint8Array.fromBase64(base64)
    return bytes.buffer
  } catch {
    // Fallback for browsers without Uint8Array.fromBase64()
    const binaryString = atob(base64)
    const bytes = new Uint8Array(binaryString.length)
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i)
    }
    return bytes.buffer
  }
}

function extractEd25519PublicKey(spkiBytes: Uint8Array): Uint8Array {
  if (spkiBytes.length !== 44) {
    throw new Error('Invalid SPKI length for Ed25519')
  }
  return spkiBytes.slice(-32)
}

async function importPublicKeyWebCrypto(publicKey: string): Promise<CryptoKey> {
  const publicKeyBuffer = base64ToArrayBuffer(publicKey)

  return await crypto.subtle.importKey(
    'spki',
    publicKeyBuffer,
    {
      name: 'Ed25519',
    },
    false,
    ['verify'],
  )
}

async function verifyWithWebCrypto(
  publicKey: string,
  scriptBuffer: ArrayBuffer,
  signatureBuffer: ArrayBuffer,
): Promise<boolean> {
  try {
    const cryptoPublicKey = await importPublicKeyWebCrypto(publicKey)

    return await crypto.subtle.verify('Ed25519', cryptoPublicKey, signatureBuffer, scriptBuffer)
  } catch (error) {
    console.log('Web Crypto verification failed:', error instanceof Error ? error.message : 'Unknown error')
    return false
  }
}

function parseSignature(signatureBuffer: ArrayBuffer): Uint8Array {
  const signatureBytes = new Uint8Array(signatureBuffer)

  // If already 64 bytes, assume it's raw binary
  if (signatureBytes.length === 64) {
    return signatureBytes
  }

  // Try to parse as text (base64 or hex)
  const signatureText = new TextDecoder().decode(signatureBytes).trim()

  // base64 decoding
  try {
    const base64Decoded = new Uint8Array(base64ToArrayBuffer(signatureText))
    if (base64Decoded.length === 64) {
      return base64Decoded
    }
  } catch (e) {
    console.error(e)
  }

  // hex decoding
  if (signatureText.length === 128 && /^[0-9a-fA-F]+$/.test(signatureText)) {
    try {
      // @ts-expect-error - fromHex is not yet in all TypeScript lib definitions
      return Uint8Array.fromHex(signatureText)
    } catch {
      // Fallback for browsers without Uint8Array.fromHex()
      const hexDecoded = new Uint8Array(64)
      for (let i = 0; i < 64; i++) {
        hexDecoded[i] = parseInt(signatureText.slice(i * 2, i * 2 + 2), 16)
      }
      return hexDecoded
    }
  }

  throw new Error(`Unable to parse signature format.`)
}

async function verifyWithNoble(
  publicKey: string,
  scriptBuffer: ArrayBuffer,
  signatureBuffer: ArrayBuffer,
): Promise<boolean> {
  try {
    const publicKeySpki = new Uint8Array(base64ToArrayBuffer(publicKey))
    const publicKeyRaw = extractEd25519PublicKey(publicKeySpki)

    const scriptBytes = new Uint8Array(scriptBuffer)
    const signatureBytes = parseSignature(signatureBuffer)

    return verify(signatureBytes, scriptBytes, publicKeyRaw)
  } catch (error) {
    console.log('Noble verification failed:', error instanceof Error ? error.message : 'Unknown error')
    return false
  }
}

self.onmessage = async function (event) {
  const { requestId, scriptUrl, signatureUrl, publicKey } = event.data

  try {
    const [scriptResponse, signatureResponse] = await Promise.all([fetch(scriptUrl), fetch(signatureUrl)])

    if (!scriptResponse.ok || !signatureResponse.ok) {
      self.postMessage({
        requestId,
        success: false,
        verified: false,
        error: `Failed to fetch files. Script: ${scriptResponse.status} ${scriptResponse.statusText}, Signature: ${signatureResponse.status} ${signatureResponse.statusText}`,
      })
      return
    }

    const [scriptBuffer, signatureBuffer] = await Promise.all([
      scriptResponse.arrayBuffer(),
      signatureResponse.arrayBuffer(),
    ])

    // Try Web Crypto API first, fallback to Noble if it fails
    let verified = await verifyWithWebCrypto(publicKey, scriptBuffer, signatureBuffer)

    if (!verified) {
      verified = await verifyWithNoble(publicKey, scriptBuffer, signatureBuffer)
    }

    // If verified, include script content to avoid re-downloading
    let scriptContent: ArrayBuffer | undefined
    if (verified) {
      scriptContent = scriptBuffer
    }

    // Send message with transferable ArrayBuffer for efficiency
    const message = {
      requestId,
      success: true,
      verified,
      scriptSize: scriptBuffer.byteLength,
      signatureSize: signatureBuffer.byteLength,
      scriptContent,
    }

    if (scriptContent) {
      self.postMessage(message, { transfer: [scriptContent] })
    } else {
      self.postMessage(message)
    }
  } catch (error) {
    console.error('[Worker] Verification error:', error)
    self.postMessage({
      requestId,
      success: false,
      verified: false,
      error: error instanceof Error ? error.message : 'Unknown error',
    })
  }
}

self.onerror = function (error) {
  console.error('[Worker] Worker error:', error)
  self.postMessage({
    success: false,
    verified: false,
    error,
  })
}
