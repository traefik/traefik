import { APP } from '../_helpers/APP'

const apiBase = ''

function getOverview () {
  return APP.api.get(`${apiBase}/overview`)
    .then(body => {
      console.log('Success -> CoreService -> getOverview', body.data)
      return body.data
    })
}

function getVersion () {
  return APP.api.get(`${apiBase}/version`)
    .then(body => {
      console.log('Success -> CoreService -> getVersion', body.data)
      return body.data
    })
}

export default {
  getOverview,
  getVersion
}
