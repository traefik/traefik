export default {
  namespaced: true,
  getters: {
    isOpen (state) {
      return state.isOpen
    }
  },
  mutations: {
    toggle (state, isOpen) {
      state.isOpen = isOpen || !state.isOpen
    }
  },
  actions: {
    toggle ({ commit }) {
      commit('toggle')
    },
    open ({ commit }) {
      commit('toggle', true)
    },
    close ({ commit }) {
      commit('toggle', false)
    }
  },
  state: {
    isOpen: false
  }
}
