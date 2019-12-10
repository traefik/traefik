import { APP } from '../_helpers/APP'

const apiBase = '/tcp'

function getAllRouters (params) {
  return APP.api.get(`${apiBase}/routers?search=${params.query}&status=${params.status}&per_page=${params.limit}&page=${params.page}`)
    .then(response => {
      const { data = [], headers } = response
      const nextPage = parseInt(headers['x-next-page'], 10) || 1
      const hasNextPage = nextPage > 1
      const total = hasNextPage
        ? (params.page + 1) * params.limit
        : params.page * params.limit
      console.log('Success -> HttpService -> getAllRouters', response.data)
      return { data, total }
    })
}

function getRouterByName (name) {
  return APP.api.get(`${apiBase}/routers/${name}`)
    .then(body => {
      console.log('Success -> HttpService -> getRouterByName', body.data)
      return body.data
    })
}

function getAllServices (params) {
  return APP.api.get(`${apiBase}/services?search=${params.query}&status=${params.status}&per_page=${params.limit}&page=${params.page}`)
    .then(response => {
      const { data = [], headers } = response
      const nextPage = parseInt(headers['x-next-page'], 10) || 1
      const hasNextPage = nextPage > 1
      const total = hasNextPage
        ? (params.page + 1) * params.limit
        : params.page * params.limit
      console.log('Success -> HttpService -> getAllServices', response.data)
      return { data, total }
    })
}

function getServiceByName (name) {
  return APP.api.get(`${apiBase}/services/${name}`)
    .then(body => {
      console.log('Success -> HttpService -> getServiceByName', body.data)
      return body.data
    })
}

export default {
  getAllRouters,
  getRouterByName,
  getAllServices,
  getServiceByName
}
