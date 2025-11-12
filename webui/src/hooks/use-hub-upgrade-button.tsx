import { useContext, useEffect, useState } from 'react'

import { VersionContext } from 'contexts/version'
import verifySignature from 'utils/workers/scriptVerification'

const HUB_BUTTON_URL = 'https://traefik.github.io/traefiklabs-hub-button-app/main-v1.js'
const PUBLIC_KEY = 'MCowBQYDK2VwAyEAY0OZFFE5kSuqYK6/UprTL5RmvQ+8dpPTGMCw1MiO/Gs='

const useHubUpgradeButton = () => {
  const [signatureVerified, setSignatureVerified] = useState(false)
  const [scriptBlobUrl, setScriptBlobUrl] = useState<string | null>(null)
  const [isCustomElementDefined, setIsCustomElementDefined] = useState(false)

  const { showHubButton } = useContext(VersionContext)

  useEffect(() => {
    if (showHubButton) {
      if (customElements.get('hub-button-app')) {
        setSignatureVerified(true)
        setIsCustomElementDefined(true)
        return
      }

      const verifyAndLoadScript = async () => {
        try {
          const { verified, scriptContent: content } = await verifySignature(
            HUB_BUTTON_URL,
            `${HUB_BUTTON_URL}.sig`,
            PUBLIC_KEY,
          )
          if (!verified || !content) {
            setSignatureVerified(false)
          } else {
            const blob = new Blob([content], { type: 'application/javascript' })
            const blobUrl = URL.createObjectURL(blob)

            setScriptBlobUrl(blobUrl)
            setSignatureVerified(true)
          }
        } catch {
          setSignatureVerified(false)
        }
      }

      verifyAndLoadScript()

      return () => {
        setScriptBlobUrl((currentUrl) => {
          if (currentUrl) {
            URL.revokeObjectURL(currentUrl)
          }
          return null
        })
      }
    }
  }, [showHubButton])

  return { signatureVerified, scriptBlobUrl, isCustomElementDefined }
}

export default useHubUpgradeButton
