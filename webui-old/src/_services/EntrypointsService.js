import { APP } from '../_helpers/APP'

const apiBase = '/entrypoints'

function getAll () {
  return APP.api.get(`${apiBase}`)
    .then(body => {
      console.log('Success -> EntrypointsService -> getAll', body.data)
      return body.data
    })
}

function getByName (name) {
  return APP.api.get(`${apiBase}/${name}`)
    .then(body => {
      console.log('Success -> EntrypointsService -> getByName', body.data)
      return body.data
    })
}

export default {
  getAll,
  getByName
}
