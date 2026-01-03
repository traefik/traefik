import { useParams } from 'react-router-dom'

import { MiddlewareDetail } from 'components/middlewares/MiddlewareDetail'
import { useResourceDetail } from 'hooks/use-resource-detail'

export const TcpMiddleware = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'middlewares', 'tcp')
  return <MiddlewareDetail data={data} error={error} name={name!} protocol="tcp" />
}

export default TcpMiddleware
