export default {
  namespaced: true,
  getters: {
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
    },
    toggleNotifVisibility (state, isHidden) {
      state.notificationIsHidden = isHidden || !state.isHidden
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
    },
    hideNotification ({ commit }) {
      commit('toggleNotifVisibility', true)
    }
  },
  state: {
    isOpen: false,
    notificationIsHidden: false
  }
}
