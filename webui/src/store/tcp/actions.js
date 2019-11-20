import TcpService from '../../_services/TcpService'

export function getAllRouters ({ commit }, params) {
  commit('getAllRoutersRequest')
  return TcpService.getAllRouters(params)
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
  return TcpService.getRouterByName(name)
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
  return TcpService.getAllServices(params)
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
  return TcpService.getServiceByName(name)
    .then(body => {
      commit('getServiceByNameSuccess', body)
      return body
    })
    .catch(error => {
      commit('getServiceByNameFailure', error)
      return Promise.reject(error)
    })
}
