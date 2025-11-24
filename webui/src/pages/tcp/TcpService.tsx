import { useParams } from 'react-router-dom'

import { ServiceDetail } from 'components/services/ServiceDetail'
import { useResourceDetail } from 'hooks/use-resource-detail'
import { NotFound } from 'pages/NotFound'

export const TcpService = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'services', 'tcp')

  if (!name) {
    return <NotFound />
  }

  return <ServiceDetail data={data} error={error} name={name} protocol="tcp" />
}

export default TcpService
