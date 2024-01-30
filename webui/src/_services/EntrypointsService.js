import { api } from 'boot/api'

const apiBase = '/entrypoints'

function getAll () {
  return api.get(`${apiBase}`)
    .then(body => {
      console.log('Success -> EntrypointsService -> getAll', body.data)
      return body.data
    })
}

function getByName (name) {
  return api.get(`${apiBase}/${name}`)
    .then(body => {
      console.log('Success -> EntrypointsService -> getByName', body.data)
      return body.data
    })
}

export default {
  getAll,
  getByName
}
