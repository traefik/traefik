import HttpService from '../../_services/HttpService'

export function getAllRouters ({ commit }, params) {
  commit('getAllRoutersRequest')
  return HttpService.getAllRouters(params)
    .then(body => {
      commit('getAllRoutersSuccess', { body, ...params })
      return body
    })
    .catch(error => {
      commit('getAllRoutersFailure', error)
      return Promise.reject(error)
    })
}

export function getRouterByName ({ commit }, name) {
  commit('getRouterByNameRequest')
  return HttpService.getRouterByName(name)
    .then(body => {
      commit('getRouterByNameSuccess', body)
      return body
    })
    .catch(error => {
      commit('getRouterByNameFailure', error)
      return Promise.reject(error)
    })
}

export function getAllServices ({ commit }, params) {
  commit('getAllServicesRequest')
  return HttpService.getAllServices(params)
    .then(body => {
      commit('getAllServicesSuccess', { body, ...params })
      return body
    })
    .catch(error => {
      commit('getAllServicesFailure', error)
      return Promise.reject(error)
    })
}

export function getServiceByName ({ commit }, name) {
  commit('getServiceByNameRequest')
  return HttpService.getServiceByName(name)
    .then(body => {
      commit('getServiceByNameSuccess', body)
      return body
    })
    .catch(error => {
      commit('getServiceByNameFailure', error)
      return Promise.reject(error)
    })
}

export function getAllMiddlewares ({ commit }, params) {
  commit('getAllMiddlewaresRequest')
  return HttpService.getAllMiddlewares(params)
    .then(body => {
      commit('getAllMiddlewaresSuccess', { body, ...params })
      return body
    })
    .catch(error => {
      commit('getAllMiddlewaresFailure', error)
      return Promise.reject(error)
    })
}

export function getMiddlewareByName ({ commit }, name) {
  commit('getMiddlewareByNameRequest')
  return HttpService.getMiddlewareByName(name)
    .then(body => {
      commit('getMiddlewareByNameSuccess', body)
      return body
    })
    .catch(error => {
      commit('getMiddlewareByNameFailure', error)
      return Promise.reject(error)
    })
}
