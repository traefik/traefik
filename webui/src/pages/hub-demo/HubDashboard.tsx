import { useMemo } from 'react'
import { Helmet } from 'react-helmet-async'
import { useParams } from 'react-router-dom'

import { useIsDarkMode } from 'hooks/use-theme'
import Page from 'layout/Page'

const HubDashboard = ({ path }: { path: string }) => {
  const isDarkMode = useIsDarkMode()

  const { id } = useParams()

  const usedPath = useMemo(() => {
    if (path?.includes(':id')) {
      const splitted = path.split(':')
      return `${splitted[0]}/${id}`
    }

    return path
  }, [id, path])

  return (
    <>
      <Helmet>
        <meta
          httpEquiv="Content-Security-Policy"
          content="script-src 'self' https://traefik.github.io 'unsafe-inline'; object-src 'none'; base-uri 'self';"
        />

        <script
          src="https://traefik.github.io/hub-ui-demo-app/scripts/hub-ui-demo.umd.js"
          type="module"
          crossOrigin="anonymous"
          referrerPolicy="strict-origin-when-cross-origin"
        ></script>
      </Helmet>

      <Page title="Hub Demo" isDemoContent>
        <hub-ui-demo-app
          key={usedPath}
          path={usedPath}
          theme={isDarkMode ? 'dark' : 'light'}
          baseurl="#/hub-dashboard"
        ></hub-ui-demo-app>
      </Page>
    </>
  )
}

export default HubDashboard
