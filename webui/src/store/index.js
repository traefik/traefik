import { createStore } from 'vuex'
import { store } from 'quasar/wrappers'

import core from './core'
import entrypoints from './entrypoints'
import http from './http'
import tcp from './tcp'
import udp from './udp'
import platform from './platform'

/*
 * If not building with SSR mode, you can
 * directly export the Store instantiation
 */

export default store((/* { ssrContext } */) => {
  const Store = createStore({
    modules: {
      core,
      entrypoints,
      http,
      tcp,
      udp,
      platform
    },

    // enable strict mode (adds overhead!)
    // for dev mode only
    strict: process.env.DEV
  })

  return Store
})
