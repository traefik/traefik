import EntrypointsService from '../../_services/EntrypointsService'

export function getAll ({ commit }) {
  commit('getAllRequest')
  return EntrypointsService.getAll()
    .then(body => {
      commit('getAllSuccess', body)
      return body
    })
    .catch(error => {
      commit('getAllFailure', error)
      return Promise.reject(error)
    })
}

export function getByName ({ commit }, name) {
  commit('getByNameRequest')
  return EntrypointsService.getByName(name)
    .then(body => {
      commit('getByNameSuccess', body)
      return body
    })
    .catch(error => {
      commit('getByNameFailure', error)
      return Promise.reject(error)
    })
}
