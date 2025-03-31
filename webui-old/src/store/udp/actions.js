import UdpService from '../../_services/UdpService'

export function getAllRouters ({ commit }, params) {
  commit('getAllRoutersRequest')
  return UdpService.getAllRouters(params)
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
  return UdpService.getRouterByName(name)
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
  return UdpService.getAllServices(params)
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
  return UdpService.getServiceByName(name)
    .then(body => {
      commit('getServiceByNameSuccess', body)
      return body
    })
    .catch(error => {
      commit('getServiceByNameFailure', error)
      return Promise.reject(error)
    })
}
