import { globalCss, Box, darkTheme, FaencyProvider, lightTheme } from '@traefiklabs/faency'
import { Suspense, useContext, useEffect } from 'react'
import { HelmetProvider } from 'react-helmet-async'
import { HashRouter, Navigate, Route, Routes as RouterRoutes, useLocation } from 'react-router-dom'
import { SWRConfig } from 'swr'

import Page from './layout/Page'
import fetch from './libs/fetch'

import { VersionProvider } from 'contexts/version'
import { useIsDarkMode } from 'hooks/use-theme'
import ErrorSuspenseWrapper from 'layout/ErrorSuspenseWrapper'
import { Dashboard, HTTPPages, NotFound, TCPPages, UDPPages } from 'pages'
import { DashboardSkeleton } from 'pages/dashboard/Dashboard'
import { HubDemoContext, HubDemoProvider } from 'pages/hub-demo/demoNavContext'

export const LIGHT_THEME = lightTheme('blue')
export const DARK_THEME = darkTheme('blue')

// TODO: Restore the loader.
const PageLoader = () => (
  <Page>
    <Box css={{ position: 'absolute', top: 0, left: 0, right: 0 }}>{/*<Loading />*/}</Box>
  </Page>
)

const ScrollToTop = () => {
  const { pathname } = useLocation()
  useEffect(() => {
    window.scrollTo(0, 0)
  }, [pathname])

  return null
}

export const Routes = () => {
  const { routes: hubDemoRoutes } = useContext(HubDemoContext)

  return (
    <Page>
      <Suspense fallback={<PageLoader />}>
        <RouterRoutes>
          <Route
            path="/"
            element={
              <ErrorSuspenseWrapper suspenseFallback={<DashboardSkeleton />}>
                <Dashboard />
              </ErrorSuspenseWrapper>
            }
          />
          <Route path="/http/routers" element={<HTTPPages.HttpRouters />} />
          <Route path="/http/services" element={<HTTPPages.HttpServices />} />
          <Route path="/http/middlewares" element={<HTTPPages.HttpMiddlewares />} />
          <Route path="/tcp/routers" element={<TCPPages.TcpRouters />} />
          <Route path="/tcp/services" element={<TCPPages.TcpServices />} />
          <Route path="/tcp/middlewares" element={<TCPPages.TcpMiddlewares />} />
          <Route path="/udp/routers" element={<UDPPages.UdpRouters />} />
          <Route path="/udp/services" element={<UDPPages.UdpServices />} />
          <Route path="/http/routers/:name" element={<HTTPPages.HttpRouter />} />
          <Route path="/http/services/:name" element={<HTTPPages.HttpService />} />
          <Route path="/http/middlewares/:name" element={<HTTPPages.HttpMiddleware />} />
          <Route path="/tcp/routers/:name" element={<TCPPages.TcpRouter />} />
          <Route path="/tcp/services/:name" element={<TCPPages.TcpService />} />
          <Route path="/tcp/middlewares/:name" element={<TCPPages.TcpMiddleware />} />
          <Route path="/udp/routers/:name" element={<UDPPages.UdpRouter />} />
          <Route path="/udp/services/:name" element={<UDPPages.UdpService />} />
          <Route path="/http" element={<Navigate to="/http/routers" replace />} />
          <Route path="/tcp" element={<Navigate to="/tcp/routers" replace />} />
          <Route path="/udp" element={<Navigate to="/udp/routers" replace />} />

          {/* Hub Dashboard demo content */}
          {hubDemoRoutes?.map((route, idx) => <Route key={`hub-${idx}`} path={route.path} element={route.element} />)}

          <Route path="*" element={<NotFound />} />
        </RouterRoutes>
      </Suspense>
    </Page>
  )
}

const isDev = import.meta.env.NODE_ENV === 'development'

const customGlobalStyle = globalCss({
  // target the AriaTd component, but exclude anything inside hub-ui-demo-app
  'body:not(:has(hub-ui-demo-app)) span[role=cell]': {
    p: '$2 $3',
  },
})

const App = () => {
  const isDarkMode = useIsDarkMode()

  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.remove(LIGHT_THEME)
      document.documentElement.classList.add(DARK_THEME)
    } else {
      document.documentElement.classList.remove(DARK_THEME)
      document.documentElement.classList.add(LIGHT_THEME)
    }
  }, [isDarkMode])

  return (
    <FaencyProvider>
      <HelmetProvider>
        <SWRConfig
          value={{
            revalidateOnFocus: !isDev,
            fetcher: fetch,
          }}
        >
          <VersionProvider>
            <HashRouter basename={import.meta.env.VITE_APP_BASE_URL || ''}>
              <HubDemoProvider basePath={'/hub-dashboard'}>
                {customGlobalStyle()}
                <ScrollToTop />
                <Routes />
              </HubDemoProvider>
            </HashRouter>
          </VersionProvider>
        </SWRConfig>
      </HelmetProvider>
    </FaencyProvider>
  )
}

export default App
