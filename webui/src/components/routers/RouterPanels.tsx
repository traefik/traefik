import { Flex, styled } from '@traefiklabs/faency'
import { useContext, useEffect, useMemo } from 'react'
import { FiGlobe, FiLayers, FiLogIn, FiZap } from 'react-icons/fi'

import { CardListSection } from 'components/resources/DetailSections'
import MiddlewarePanel from 'components/resources/MiddlewarePanel'
import RouterPanel from 'components/resources/RouterPanel'
import TlsPanel from 'components/resources/TlsPanel'
import { ToastContext } from 'contexts/toasts'
import { EntryPoint, ResourceDetailDataType } from 'hooks/use-resource-detail'
import { getErrorData, getValidData } from 'libs/objectHandlers'
import { parseMiddlewareType } from 'libs/parsers'

type RouterPanelsProps = {
  data: ResourceDetailDataType
  protocol: 'http' | 'tcp' | 'udp'
}

const CardListColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(4, 1fr)',
  marginBottom: '48px',
})

const SpacedColumns = styled(Flex, {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))',
  gridGap: '16px',
})

export const RouterPanels = ({ data, protocol }: RouterPanelsProps) => {
  const { addToast } = useContext(ToastContext)
  const entrypoints = useMemo(() => getValidData(data.entryPointsData), [data?.entryPointsData])
  const entrypointsError = useMemo(() => getErrorData(data.entryPointsData), [data?.entryPointsData])

  const serviceSlug = data.service?.includes('@')
    ? data.service
    : `${data.service ?? 'unknown'}@${data.provider ?? 'unknown'}`

  const isUdp = useMemo(() => protocol === 'udp', [protocol])

  useEffect(() => {
    entrypointsError?.map((error) =>
      addToast({
        message: error.message,
        severity: 'error',
      }),
    )
  }, [addToast, entrypointsError])

  return (
    <>
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

      <SpacedColumns data-testid="router-details">
        <RouterPanel data={data} />
        {!isUdp && <TlsPanel data={data} />}
        {!isUdp && <MiddlewarePanel data={data} />}
      </SpacedColumns>
    </>
  )
}
