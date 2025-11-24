import { useParams } from 'react-router-dom'

import { RouterDetail } from 'components/routers/RouterDetail'
import { useResourceDetail } from 'hooks/use-resource-detail'
import { NotFound } from 'pages/NotFound'

export const TcpRouter = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name!, 'routers', 'tcp')

  if (!name) {
    return <NotFound />
  }

  return <RouterDetail data={data} error={error} name={name} protocol="tcp" />
}

export default TcpRouter
