import { Box, Flex, Image, Link, Text } from '@traefiklabs/faency'
import { useMemo, useEffect, useState } from 'react'
import { Helmet } from 'react-helmet-async'
import { useParams } from 'react-router-dom'

import { verifyScriptSignature } from './workers/scriptVerification'

import { SpinnerLoader } from 'components/SpinnerLoader'
import { useIsDarkMode } from 'hooks/use-theme'
import { TopNav } from 'layout/Navigation'

const PUBLIC_KEY = 'MCowBQYDK2VwAyEAWMBZ0pMBaL/s8gNXxpAPCIQ8bxjnuz6bQFwGYvjXDfg='

const HubDashboard = ({ path }: { path: string }) => {
  const isDarkMode = useIsDarkMode()
  const [scriptError, setScriptError] = useState<string | null>(null)
  const [signatureVerified, setSignatureVerified] = useState(false)
  const [verificationInProgress, setVerificationInProgress] = useState(false)
  const [scriptBlobUrl, setScriptBlobUrl] = useState<string | null>(null)

  const { id } = useParams()

  const usedPath = useMemo(() => {
    if (path?.includes(':id')) {
      const splitted = path.split(':')
      return `${splitted[0]}/${id}`
    }

    return path
  }, [id, path])

  useEffect(() => {
    const verifyAndLoadScript = async () => {
      setVerificationInProgress(true)

      try {
        const scriptPath = 'https://assets.traefik.io/hub-ui-demo.js'
        const signaturePath = 'https://assets.traefik.io/hub-ui-demo.js.sig'

        const result = await verifyScriptSignature(PUBLIC_KEY, scriptPath, signaturePath)

        if (!result.verified || !result.scriptContent) {
          setScriptError('Script signature verification failed - security violation detected')
          setVerificationInProgress(false)
          return
        }

        // Create Blob URL from verified script content
        const blob = new Blob([result.scriptContent], { type: 'application/javascript' })
        const blobUrl = URL.createObjectURL(blob)

        setScriptBlobUrl(blobUrl)
        setSignatureVerified(true)
        setVerificationInProgress(false)
      } catch (error) {
        setScriptError(`Script verification failed: ${error instanceof Error ? error.message : 'Unknown error'}`)
        setVerificationInProgress(false)
      }
    }

    verifyAndLoadScript()

    // Cleanup: revoke Blob URL on unmount
    return () => {
      if (scriptBlobUrl) {
        URL.revokeObjectURL(scriptBlobUrl)
      }
    }
  }, [])

  if (scriptError) {
    return (
      <Flex gap={4} align="center" justify="center" direction="column" css={{ width: '100%', mt: '$8', maxWidth: 690 }}>
        <Image src="/img/gopher-something-went-wrong.png" width={400} />
        <Text css={{ fontSize: 24, fontWeight: '$semiBold' }}>Oops! We couldn't load the demo content.</Text>
        <Text size={6} css={{ textAlign: 'center', lineHeight: 1.4 }}>
          Don't worry — you can still learn more about{' '}
          <Text size={6} css={{ fontWeight: '$semiBold' }}>
            Traefik Hub API Management
          </Text>{' '}
          on our{' '}
          <Link
            href="https://traefik.io/traefik-hub?utm_campaign=hub-demo&utm_source=proxy-button&utm_medium=in-product"
            target="_blank"
          >
            website
          </Link>{' '}
          or in our{' '}
          <Link href="https://doc.traefik.io/traefik-hub/" target="_blank">
            documentation
          </Link>
          .
        </Text>
      </Flex>
    )
  }

  return (
    <Box css={{ width: '100%' }}>
      <Helmet>
        <title>Hub Demo - Traefik Proxy</title>
        <meta
          httpEquiv="Content-Security-Policy"
          content="script-src 'self' blob: 'unsafe-inline'; object-src 'none'; base-uri 'self';"
        />
        {signatureVerified && scriptBlobUrl && <script src={scriptBlobUrl} type="module"></script>}
      </Helmet>
      <Box
        css={{
          margin: 'auto',
          position: 'relative',
          maxWidth: '1334px',
          '@media (max-width:1440px)': {
            maxWidth: '100%',
          },
        }}
      >
        <TopNav noHubButton css={{ position: 'absolute', top: 125, right: '$5' }} />
      </Box>
      {verificationInProgress ? (
        <Box css={{ width: '100%', justifyItems: 'center', mt: '$8' }}>
          <SpinnerLoader size={48} />
        </Box>
      ) : (
        <hub-ui-demo-app
          key={usedPath}
          path={usedPath}
          theme={isDarkMode ? 'dark' : 'light'}
          baseurl="#/hub-dashboard"
          containercss={JSON.stringify({
            maxWidth: '1334px',
            '@media (max-width:1440px)': {
              maxWidth: '100%',
            },
            margin: 'auto',
            marginTop: '90px',
          })}
        ></hub-ui-demo-app>
      )}
    </Box>
  )
}

export default HubDashboard
