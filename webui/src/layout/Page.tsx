import { Flex, globalCss, styled } from '@traefiklabs/faency'
import { ReactNode, useState } from 'react'
import { Helmet } from 'react-helmet-async'

import Container from './Container'
import { SideBarPanel, SideNav, TopNav } from './NavBar'

import { ToastPool } from 'components/ToastPool'
import { ToastProvider } from 'contexts/toasts'

export const globalStyles = globalCss({
  body: {
    backgroundColor: '$contentBg',
    m: 0,
  },
})

const PageContainer = styled(Container, {
  py: '$5',
  px: '$5',
  m: 0,
})

export interface Props {
  title?: string
  children?: ReactNode
}

const Page = ({ children, title }: Props) => {
  const [isSideBarPanelOpen, setIsSideBarPanelOpen] = useState(false)

  return (
    <ToastProvider>
      {globalStyles()}
      <Helmet>
        <title>{title ? `${title} - ` : ''}Traefik Proxy</title>
      </Helmet>
      <Flex>
        <SideBarPanel isOpen={isSideBarPanelOpen} onOpenChange={setIsSideBarPanelOpen} />
        <SideNav isExpanded={isSideBarPanelOpen} onSidePanelToggle={() => setIsSideBarPanelOpen(true)} isResponsive />
        <Flex justify="center" css={{ flex: 1, height: '100vh', overflowY: 'auto', margin: 'auto' }}>
          <PageContainer data-testid={`${title} page`} direction="column">
            <TopNav />
            {children}
          </PageContainer>
        </Flex>
      </Flex>
      <ToastPool />
    </ToastProvider>
  )
}

export default Page
