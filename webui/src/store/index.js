import Vue from 'vue'
import Vuex from 'vuex'

import core from './core'
import entrypoints from './entrypoints'
import http from './http'
import tcp from './tcp'
import udp from './udp'
import platform from './platform'

Vue.use(Vuex)

/*
 * If not building with SSR mode, you can
 * directly export the Store instantiation
 */

export default function (/* { ssrContext } */) {
  const Store = new Vuex.Store({
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
}
