import { APP } from '../_helpers/APP'

const apiBase = '/http'

function getAllRouters (params) {
  return APP.api.get(`${apiBase}/routers?search=${params.query}&status=${params.status}&per_page=${params.limit}&page=${params.page}`)
    .then(body => {
      const total = body.data ? body.data.length : 0
      console.log('Success -> HttpService -> getAllRouters', body.data)
      // TODO - suggestion: add the total-pages in api response to optimize the query
      return { data: body.data || [], total }
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
    .then(body => {
      const total = body.data ? body.data.length : 0
      console.log('Success -> HttpService -> getAllServices', body.data)
      // TODO - suggestion: add the total-pages in api response to optimize the query
      return { data: body.data || [], total }
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
    .then(body => {
      const total = body.data ? body.data.length : 0
      console.log('Success -> HttpService -> getAllMiddlewares', body.data)
      // TODO - suggestion: add the total-pages in api response to optimize the query
      return { data: body.data || [], total }
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
