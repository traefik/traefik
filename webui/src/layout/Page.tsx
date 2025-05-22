import { Flex, globalCss, styled } from '@traefiklabs/faency'
import { ReactNode } from 'react'
import { Helmet } from 'react-helmet-async'

import Container from './Container'
import { SideNav, TopNav } from './NavBar'

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
  height: '100vh',
  overflowY: 'auto',
  m: 0,
})

export interface Props {
  title?: string
  children?: ReactNode
}

const Page = ({ children, title }: Props) => {
  return (
    <ToastProvider>
      {globalStyles()}
      <Helmet>
        <title>{title ? `${title} - ` : ''}Traefik Proxy</title>
      </Helmet>
      <Flex>
        <SideNav />
        <PageContainer data-testid={`${title} page`} direction="column">
          <TopNav />
          {children}
        </PageContainer>
      </Flex>
      <ToastPool />
    </ToastProvider>
  )
}

export default Page
