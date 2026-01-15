import { Flex, globalCss, styled } from '@traefiklabs/faency'
import { ReactNode, useMemo, useState } from 'react'
import { useLocation } from 'react-router-dom'

import Container from './Container'
import PageTitle from './PageTitle'

import { ToastPool } from 'components/ToastPool'
import { ToastProvider } from 'contexts/toasts'
import { LAPTOP_BP, SideBarPanel, SideNav, TopNav } from 'layout/navigation'

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

const Page = ({ children }: Props) => {
  const { pathname } = useLocation()
  const [isSideBarPanelOpen, setIsSideBarPanelOpen] = useState(false)
  const location = useLocation()

  const isDemoPage = useMemo(() => pathname.includes('hub-dashboard'), [pathname])

  const renderedContent = useMemo(() => {
    if (isDemoPage) {
      return children
    }

    return (
      <PageContainer data-testid={`${location.pathname} page`} direction="column">
        <TopNav />
        {children}
      </PageContainer>
    )
  }, [children, isDemoPage, location.pathname])

  return (
    <ToastProvider>
      {globalStyles()}
      <PageTitle />
      <Flex>
        <SideBarPanel isOpen={isSideBarPanelOpen} onOpenChange={setIsSideBarPanelOpen} />
        <SideNav isExpanded={isSideBarPanelOpen} onSidePanelToggle={() => setIsSideBarPanelOpen(true)} isResponsive />
        <Flex
          justify="center"
          css={{ flex: 1, margin: 'auto', ml: 264, [`@media (max-width:${LAPTOP_BP}px)`]: { ml: 60 } }}
        >
          {renderedContent}
        </Flex>
      </Flex>
      <ToastPool />
    </ToastProvider>
  )
}

export default Page
