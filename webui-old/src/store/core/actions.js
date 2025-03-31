import coreService from '../../_services/CoreService'

export function getOverview ({ commit }) {
  commit('getOverviewRequest')
  return coreService.getOverview()
    .then(body => {
      commit('getOverviewSuccess', body)
      return body
    })
    .catch(error => {
      commit('getOverviewFailure', error)
      return Promise.reject(error)
    })
}

export function getVersion ({ commit }) {
  return coreService.getVersion()
    .then(body => {
      commit('getVersionSuccess', body)
      return body
    })
    .catch(error => {
      return Promise.reject(error)
    })
}
