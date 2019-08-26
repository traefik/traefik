import entrypointsService from '../../_services/EntrypointsService'

export function getAll ({ commit }) {
  commit('getAllRequest')
  return entrypointsService.getAll()
    .then(body => {
      commit('getAllSuccess', body)
      return body
    })
    .catch(error => {
      commit('getAllFailure', error)
      return Promise.reject(error)
    })
}
