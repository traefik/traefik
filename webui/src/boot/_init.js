import { boot } from 'quasar/wrappers'
import errors from '../_helpers/Errors'

export default boot(({ app }) => {
  // Router
  // ----------------------------------------------
  // app.$router.beforeEach(async (to, from, next) => {
  //   // Set APP
  //   APP.routeTo = to
  //   APP.routeFrom = from
  //   next()
  // })

  // Api (axios)
  // ----------------------------------------------
  app.config.globalProperties.$api.interceptors.request.use((config) => {
    console.log('interceptors -> config', config)
    // config.headers['Accept'] = '*/*'
    return config
  })

  app.config.globalProperties.$api.interceptors.response.use((response) => {
    console.log('interceptors -> response', response)
    return response
  }, errors.handleResponse)
})
