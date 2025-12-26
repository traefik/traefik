import { useParams } from 'react-router-dom'

import { MiddlewareDetail } from 'components/middlewares/MiddlewareDetail'
import { useResourceDetail } from 'hooks/use-resource-detail'

export const HttpMiddleware = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'middlewares')
  return <MiddlewareDetail data={data} error={error} name={name!} protocol="http" />
}

export default HttpMiddleware
