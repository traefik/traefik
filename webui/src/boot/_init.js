import { APP } from '../_helpers/APP'
import errors from '../_helpers/Errors'
import resize from '../_directives/resize'

export default async ({ app, router }) => {
  // Directives
  app.directive('resize', resize)

  // Router
  // ----------------------------------------------
  router.beforeEach(async (to, from, next) => {
    // Set APP
    APP.routeTo = to
    APP.routeFrom = from
    next()
  })

  // Api (axios)
  // ----------------------------------------------
  APP.api.interceptors.request.use((config) => {
    console.log('interceptors -> config', config)
    // config.headers['Accept'] = '*/*'
    return config
  })

  APP.api.interceptors.response.use((response) => {
    console.log('interceptors -> response', response)
    return response
  }, errors.handleResponse)
}
