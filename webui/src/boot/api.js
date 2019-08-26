import axios from 'axios'

import { APP } from '../_helpers/APP'

// Set config defaults when creating the instance
const API = axios.create({
  baseURL: APP.config.apiUrl
})

export default async ({ app, Vue }) => {
  Vue.prototype.$api = app.api = APP.api = API
}
