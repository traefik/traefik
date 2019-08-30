import HttpService from '../../_services/HttpService'

export function getAllRouters ({ commit }, params) {
  commit('getAllRoutersRequest')
  return HttpService.getAllRouters(params)
    .then(body => {
      commit('getAllRoutersSuccess', body)
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
