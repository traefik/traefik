import { Box, Flex, Image, Link, Text } from '@traefiklabs/faency'
import { useMemo, useEffect, useState } from 'react'
import { Helmet } from 'react-helmet-async'
import { useParams } from 'react-router-dom'

import { verifyScriptSignature } from './workers/scriptVerification'

import { SpinnerLoader } from 'components/SpinnerLoader'
import { useIsDarkMode } from 'hooks/use-theme'
import Page from 'layout/Page'

const PUBLIC_KEY = 'MCowBQYDK2VwAyEAWMBZ0pMBaL/s8gNXxpAPCIQ8bxjnuz6bQFwGYvjXDfg='

const HubDashboard = ({ path }: { path: string }) => {
  const isDarkMode = useIsDarkMode()
  const [scriptError, setScriptError] = useState<string | null>(null)
  const [signatureVerified, setSignatureVerified] = useState(false)
  const [verificationInProgress, setVerificationInProgress] = useState(false)

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
        const scriptPath = 'https://traefik.github.io/hub-ui-demo-app/scripts/hub-ui-demo.umd.js'
        const signaturePath = 'https://traefik.github.io/hub-ui-demo-app/scripts/hub-ui-demo.umd.js.sig'

        const isSignatureValid = await verifyScriptSignature(PUBLIC_KEY, scriptPath, signaturePath)

        if (!isSignatureValid) {
          setScriptError('Script signature verification failed - security violation detected')
          setVerificationInProgress(false)
          return
        }

        setSignatureVerified(true)
        setVerificationInProgress(false)
      } catch (error) {
        setScriptError(`Script verification failed: ${error instanceof Error ? error.message : 'Unknown error'}`)
        setVerificationInProgress(false)
      }
    }

    verifyAndLoadScript()
  }, [])

  if (scriptError) {
    return (
      <Page title="Hub Demo" isDemoContent>
        <Flex gap={4} align="center" justify="center" direction="column" css={{ width: '100%', mt: '$8' }}>
          <Image src="/img/gopher-something-went-wrong.png" width={400} />
          <Text css={{ fontSize: 24, fontWeight: '$semiBold' }}>
            Oops, the demo content couldn't be fetched correctly
          </Text>
          <Text size={6} css={{ textAlign: 'center', lineHeight: 1.3 }}>
            But don't worry, you can still read more about Traefik Hub API Management on our{' '}
            <Link href="https://traefik.io/traefik-hub" target="_blank">
              website
            </Link>{' '}
            and on our{' '}
            <Link href="https://doc.traefik.io/traefik-hub/" target="_blank">
              documentation
            </Link>
            .
          </Text>
        </Flex>
      </Page>
    )
  }

  return (
    <>
      <Helmet>
        <meta
          httpEquiv="Content-Security-Policy"
          content="script-src 'self' https://traefik.github.io 'unsafe-inline'; object-src 'none'; base-uri 'self';"
        />
        {signatureVerified && (
          <script
            src="https://traefik.github.io/hub-ui-demo-app/scripts/hub-ui-demo.umd.js"
            type="module"
            crossOrigin="anonymous"
            referrerPolicy="strict-origin-when-cross-origin"
          ></script>
        )}
      </Helmet>

      <Page title="Hub Demo" isDemoContent>
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
          ></hub-ui-demo-app>
        )}
      </Page>
    </>
  )
}

export default HubDashboard
