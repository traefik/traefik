import { Notify } from 'quasar'
import { APP } from './APP'

class Errors {
  // Getters
  // ------------------------------------------------------------------------

  // Public
  // ------------------------------------------------------------------------

  // Static
  // ------------------------------------------------------------------------

  static showError (body) {
    body = APP._.isString(body) ? JSON.parse(body) : body
    Notify.create({
      color: 'negative',
      position: 'top',
      message: body.message, // TODO - get correct error message
      icon: 'report_problem'
    })
  }

  static handleResponse (error) {
    console.log('handleResponse', error, error.response)
    let body = error.response.data
    if (error.response.status === 401) {
      // TODO - actions...
    }
    Errors.showError(body)
    return Promise.reject(body)
  }

  // Static Private
  // ------------------------------------------------------------------------
}

export default Errors
