import { Flex, styled, Text } from '@traefiklabs/faency'
import { useContext, useEffect, useMemo } from 'react'
import { FiGlobe, FiLayers, FiLogIn, FiZap } from 'react-icons/fi'
import { useParams } from 'react-router-dom'

import { CardListSection, DetailSectionSkeleton } from 'components/resources/DetailSections'
import MiddlewarePanel from 'components/resources/MiddlewarePanel'
import RouterPanel from 'components/resources/RouterPanel'
import TlsPanel from 'components/resources/TlsPanel'
import { ToastContext } from 'contexts/toasts'
import { EntryPoint, ResourceDetailDataType, useResourceDetail } from 'hooks/use-resource-detail'
import Page from 'layout/Page'
import { getErrorData, getValidData } from 'libs/objectHandlers'
import { parseMiddlewareType } from 'libs/parsers'
import { NotFound } from 'pages/NotFound'

const CardListColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(4, 1fr)',
  marginBottom: '48px',
})

type DetailProps = {
  data: ResourceDetailDataType
  protocol?: string
}

export const RouterStructure = ({ data, protocol = 'http' }: DetailProps) => {
  const { addToast } = useContext(ToastContext)
  const entrypoints = useMemo(() => getValidData(data.entryPointsData), [data?.entryPointsData])
  const entrypointsError = useMemo(() => getErrorData(data.entryPointsData), [data?.entryPointsData])

  const serviceSlug = data.service?.includes('@')
    ? data.service
    : `${data.service ?? 'unknown'}@${data.provider ?? 'unknown'}`

  useEffect(() => {
    entrypointsError?.map((error) =>
      addToast({
        message: error.message,
        severity: 'error',
      }),
    )
  }, [addToast, entrypointsError])

  return (
    <CardListColumns data-testid="router-structure">
      {entrypoints.length > 0 && (
        <CardListSection
          bigDescription
          icon={<FiLogIn size={20} />}
          title="Entrypoints"
          cards={data.entryPointsData?.map((ep: EntryPoint) => ({
            title: ep.name,
            description: ep.address,
          }))}
        />
      )}
      <CardListSection
        icon={<FiGlobe size={20} />}
        title={`${protocol.toUpperCase()} Router`}
        cards={[{ title: 'router', description: data.name, focus: true }]}
      />
      {data.hasValidMiddlewares && (
        <CardListSection
          icon={<FiLayers size={20} />}
          title={`${protocol.toUpperCase()} Middlewares`}
          cards={data.middlewares?.map((mw) => ({
            title: parseMiddlewareType(mw) ?? 'middleware',
            description: mw.name,
            link: `/${protocol}/middlewares/${mw.name}`,
          }))}
        />
      )}
      <CardListSection
        isLast
        icon={<FiZap size={20} />}
        title="Service"
        cards={[{ title: 'service', description: data.service, link: `/${protocol}/services/${serviceSlug}` }]}
      />
    </CardListColumns>
  )
}

const SpacedColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))',
  gridGap: '16px',
})

const RouterDetail = ({ data }: DetailProps) => (
  <SpacedColumns data-testid="router-detail">
    <RouterPanel data={data} />
    <TlsPanel data={data} />
    <MiddlewarePanel data={data} />
  </SpacedColumns>
)

type HttpRouterRenderProps = {
  data?: ResourceDetailDataType
  error?: Error | null
  name: string
}

export const HttpRouterRender = ({ data, error, name }: HttpRouterRenderProps) => {
  if (error) {
    return (
      <Page title={name}>
        <Text data-testid="error-text">
          Sorry, we could not fetch detail information for this Router right now. Please, try again later.
        </Text>
      </Page>
    )
  }

  if (!data) {
    return (
      <Page title={name}>
        <Flex css={{ flexDirection: 'row', mb: '70px' }} data-testid="skeleton">
          <CardListSection bigDescription />
          <CardListSection />
          <CardListSection />
          <CardListSection isLast />
        </Flex>
        <SpacedColumns>
          <DetailSectionSkeleton />
          <DetailSectionSkeleton />
          <DetailSectionSkeleton />
        </SpacedColumns>
      </Page>
    )
  }

  if (!data.name) {
    return <NotFound />
  }

  return (
    <Page title={name}>
      <RouterStructure data={data} protocol="http" />
      <RouterDetail data={data} />
    </Page>
  )
}

export const HttpRouter = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'routers')
  return <HttpRouterRender data={data} error={error} name={name!} />
}

export default HttpRouter
