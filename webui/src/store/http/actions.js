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
