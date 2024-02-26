import { boot } from 'quasar/wrappers'
import axios from 'axios'
import { APP } from '../_helpers/APP'

// Set config defaults when creating the instance
const api = axios.create({
  baseURL: APP.config.apiUrl
})

export default boot(({ app }) => {
  app.config.globalProperties.$axios = axios
  app.config.globalProperties.$api = api
  APP.api = api
})

export { api }
