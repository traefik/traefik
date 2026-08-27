import { useParams } from 'react-router'

import { RouterDetail } from 'components/routers/RouterDetail'
import { useResourceDetail } from 'hooks/use-resource-detail'

export const UdpRouter = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'routers', 'udp')

  return <RouterDetail data={data} error={error} name={name!} protocol="udp" />
}

export default UdpRouter
