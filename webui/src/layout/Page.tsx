import { globalCss, styled } from '@traefiklabs/faency'
import { ToastPool } from 'components/ToastPool'
import { ToastProvider } from 'contexts/toasts'
import { ReactNode } from 'react'
import { Helmet } from 'react-helmet-async'

import Container from './Container'
import Header from './Header'

export const globalStyles = globalCss({
  body: {
    backgroundColor: '$contentBg',
    m: 0,
  },
})

const PageContainer = styled(Container, {
  py: '$5',
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
      <Header />
      <PageContainer data-testid={`${title} page`} direction="column">
        {children}
      </PageContainer>
      <ToastPool />
    </ToastProvider>
  )
}

export default Page
