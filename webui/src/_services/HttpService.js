import { APP } from '../_helpers/APP'
import { getTotal } from './utils'

const apiBase = '/http'

function getAllRouters (params) {
  return APP.api.get(`${apiBase}/routers?search=${params.query}&status=${params.status}&per_page=${params.limit}&page=${params.page}`)
    .then(response => {
      const { data = [], headers } = response
      const total = getTotal(headers, params)
      console.log('Success -> HttpService -> getAllRouters', response, response.data)
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
      const total = getTotal(headers, params)
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

function getAllMiddlewares (params) {
  return APP.api.get(`${apiBase}/middlewares?search=${params.query}&status=${params.status}&per_page=${params.limit}&page=${params.page}`)
    .then(response => {
      const { data = [], headers } = response
      const total = getTotal(headers, params)
      console.log('Success -> HttpService -> getAllMiddlewares', response.data)
      return { data, total }
    })
}

function getMiddlewareByName (name) {
  return APP.api.get(`${apiBase}/middlewares/${name}`)
    .then(body => {
      console.log('Success -> HttpService -> getMiddlewareByName', body.data)
      return body.data
    })
}

export default {
  getAllRouters,
  getRouterByName,
  getAllServices,
  getServiceByName,
  getAllMiddlewares,
  getMiddlewareByName
}
