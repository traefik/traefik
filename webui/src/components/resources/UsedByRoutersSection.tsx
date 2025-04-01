import { Box, Flex, styled, Table, Tbody, Td, Th, Thead } from '@traefiklabs/faency'
import { useContext, useEffect, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'

import { SectionHeader } from './DetailSections'

import { Tr } from 'components/FaencyOverrides'
import { ToastContext } from 'contexts/toasts'
import { MiddlewareDetailType, ServiceDetailType } from 'hooks/use-resource-detail'
import { makeRowRender } from 'pages/http/HttpRouters'

type UsedByRoutersSectionProps = {
  data: ServiceDetailType | MiddlewareDetailType
  protocol?: string
}

const SkeletonContent = styled(Box, {
  backgroundColor: '$slate5',
  height: '14px',
  minWidth: '50px',
  borderRadius: '4px',
  margin: '8px',
})

export const UsedByRoutersSkeleton = () => (
  <Flex css={{ flexDirection: 'column', mt: '40px' }}>
    <SectionHeader />
    <Table>
      <Thead>
        <Tr>
          <Th>
            <SkeletonContent />
          </Th>
          <Th>
            <SkeletonContent />
          </Th>
          <Th>
            <SkeletonContent />
          </Th>
          <Th>
            <SkeletonContent />
          </Th>
          <Th>
            <SkeletonContent />
          </Th>
        </Tr>
      </Thead>
      <Tbody>
        <Tr style={{ pointerEvents: 'none' }}>
          <Td>
            <SkeletonContent />
          </Td>
          <Td>
            <SkeletonContent />
          </Td>
          <Td>
            <SkeletonContent />
          </Td>
          <Td>
            <SkeletonContent />
          </Td>
          <Td>
            <SkeletonContent />
          </Td>
        </Tr>
        <Tr style={{ pointerEvents: 'none' }}>
          <Td>
            <SkeletonContent />
          </Td>
          <Td>
            <SkeletonContent />
          </Td>
          <Td>
            <SkeletonContent />
          </Td>
          <Td>
            <SkeletonContent />
          </Td>
          <Td>
            <SkeletonContent />
          </Td>
        </Tr>
      </Tbody>
    </Table>
  </Flex>
)

export const UsedByRoutersSection = ({ data, protocol = 'http' }: UsedByRoutersSectionProps) => {
  const navigate = useNavigate()
  const renderRow = makeRowRender(navigate, protocol)
  const { addToast } = useContext(ToastContext)

  const routersFound = useMemo(() => data.routers?.filter((r) => !r.message), [data])
  const routersNotFound = useMemo(() => data.routers?.filter((r) => !!r.message), [data])

  useEffect(() => {
    routersNotFound?.map((error) =>
      addToast({
        message: error.message,
        severity: 'error',
      }),
    )
  }, [addToast, routersNotFound])

  if (!routersFound || routersFound.length <= 0) {
    return null
  }

  return (
    <Flex css={{ flexDirection: 'column', mt: '$5' }}>
      <SectionHeader title="Used by Routers" />

      <Box css={{ maxWidth: '100%', width: '100%', overflow: 'auto' }}>
        <Table data-testid="routers-table" css={{ tableLayout: 'auto' }}>
          <Thead>
            <Tr>
              <Th>Status</Th>
              <Th>TLS</Th>
              <Th>Rule</Th>
              <Th>Entrypoints</Th>
              <Th>Name</Th>
              <Th>Service</Th>
              <Th>Provider</Th>
            </Tr>
          </Thead>
          <Tbody>{routersFound.map(renderRow)}</Tbody>
        </Table>
      </Box>
    </Flex>
  )
}
