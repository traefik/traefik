import { useParams } from 'react-router-dom'

import { RouterDetail } from 'components/routers/RouterDetail'
import { useResourceDetail } from 'hooks/use-resource-detail'

export const TcpRouter = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'routers', 'tcp')

  return <RouterDetail data={data} error={error} name={name!} protocol="tcp" />
}

export default TcpRouter
