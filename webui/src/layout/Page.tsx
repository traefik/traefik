import { Flex, globalCss, styled } from '@traefiklabs/faency'
import { ReactNode, useState } from 'react'
import { Helmet } from 'react-helmet-async'

import Container from './Container'
import { LAPTOP_BP, SideBarPanel, SideNav, TopNav } from './Navigation'

import { ToastPool } from 'components/ToastPool'
import { ToastProvider } from 'contexts/toasts'

export const LIGHT_PRIMARY_COLOR = '#217F97'
export const DARK_PRIMARY_COLOR = '#2AA2C1'

export const globalStyles = globalCss({
  '.light': {
    '--colors-primary': LIGHT_PRIMARY_COLOR,
  },

  '.dark': {
    '--colors-primary': DARK_PRIMARY_COLOR,
  },

  body: {
    backgroundColor: '$contentBg',
    m: 0,
  },
})

const PageContainer = styled(Container, {
  py: '$5',
  px: '$5',
  m: 0,
  '@media (max-width:1440px)': {
    maxWidth: '100%',
  },
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
        <Flex
          justify="center"
          css={{ flex: 1, margin: 'auto', ml: 264, [`@media (max-width:${LAPTOP_BP}px)`]: { ml: 60 } }}
        >
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
