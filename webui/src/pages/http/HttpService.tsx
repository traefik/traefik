import { useParams } from 'react-router-dom'

import { ServiceDetail } from 'components/services/ServiceDetail'
import { useResourceDetail } from 'hooks/use-resource-detail'
import { NotFound } from 'pages/NotFound'

export const HttpService = () => {
  const { name } = useParams<{ name: string }>()
  const { data, error } = useResourceDetail(name ?? '', 'services')

  if (!name) {
    return <NotFound />
  }

  return <ServiceDetail data={data} error={error} name={name} protocol="http" />
}

export default HttpService
