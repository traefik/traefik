export default {
  namespaced: true,
  getters: {
    path (state) {
      return state.path
    },
    isOpen (state) {
      return state.isOpen
    },
    notificationIsHidden (state) {
      return state.notificationIsHidden
    }
  },
  mutations: {
    toggle (state, isOpen) {
      state.isOpen = isOpen || !state.isOpen
      if (!state.isOpen) {
        state.path = '/'
      }
    },
    setPath (state, path = '/') {
      state.path = path
    },
    toggleNotifVisibility (state, isHidden) {
      state.notificationIsHidden = isHidden || !state.isHidden
    }
  },
  actions: {
    toggle ({ commit }) {
      commit('toggle')
    },
    open ({ commit }, path) {
      commit('setPath', path)
      commit('toggle', true)
    },
    close ({ commit }) {
      commit('setPath', '/')
      commit('toggle', false)
    },
    hideNotification ({ commit }) {
      commit('toggleNotifVisibility', true)
    }
  },
  state: {
    path: '/',
    isOpen: false,
    notificationIsHidden: false
  }
}
