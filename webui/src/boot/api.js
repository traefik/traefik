import { boot } from 'quasar/wrappers'
import axios from 'axios'

// Set config defaults when creating the instance
const api = axios.create({
  baseURL: '/api'//  FIXME: I must use "APP.config.apiUrl" here
})

export default boot(({ app }) => {
  app.config.globalProperties.$axios = axios
  app.config.globalProperties.$api = api
})

export { api }
