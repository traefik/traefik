import { useParams } from 'react-router-dom'

import { ServiceDetail } from 'components/services/ServiceDetail'
import { useResourceDetail } from 'hooks/use-resource-detail'

export const UdpService = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'services', 'udp')

  return <ServiceDetail data={data} error={error} name={name!} protocol="udp" />
}

export default UdpService
