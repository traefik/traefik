import { expect } from 'chai'
import store from './index.js'

const {
  getAllRoutersRequest,
  getAllRoutersSuccess,
  getAllRoutersFailure,
  getAllServicesRequest,
  getAllServicesSuccess,
  getAllServicesFailure
} = store.mutations

describe('tcp mutations', function () {
  /* Routers */
  describe('tcp routers mutations', function () {
    it('getAllRoutersRequest', function () {
      const state = {
        allRouters: {}
      }

      getAllRoutersRequest(state)

      expect(state.allRouters.loading).to.equal(true)
    })

    it('getAllRoutersSuccess', function () {
      const state = {
        allRouters: {
          loading: true
        }
      }

      const data = {
        body: {
          data: [{}, {}, {}],
          total: 3
        },
        query: 'test query',
        status: 'warning',
        page: 1
      }

      getAllRoutersSuccess(state, data)

      expect(state.allRouters.loading).to.equal(false)
      expect(state.allRouters.total).to.equal(3)
      expect(state.allRouters.items.length).to.equal(3)
      expect(state.allRouters.currentPage).to.equal(1)
      expect(state.allRouters.currentQuery).to.equal('test query')
      expect(state.allRouters.currentStatus).to.equal('warning')
    })

    it('getAllRoutersFailing', function () {
      const state = {
        allRouters: {
          loading: true
        }
      }

      const error = { message: 'invalid request: page: 3, per_page: 10' }

      getAllRoutersFailure(state, error)

      expect(state.allRouters.loading).to.equal(false)
      expect(state.allRouters.endReached).to.equal(true)
    })
  })

  /* Services */
  describe('tcp services mutations', function () {
    it('getAllServicesRequest', function () {
      const state = {
        allServices: {}
      }

      getAllServicesRequest(state)

      expect(state.allServices.loading).to.equal(true)
    })

    it('getAllServicesSuccess', function () {
      const state = {
        allServices: {
          loading: true
        }
      }

      const data = {
        body: {
          data: [{}, {}, {}],
          total: 3
        },
        query: 'test query',
        status: 'warning',
        page: 1
      }

      getAllServicesSuccess(state, data)

      expect(state.allServices.loading).to.equal(false)
      expect(state.allServices.total).to.equal(3)
      expect(state.allServices.items.length).to.equal(3)
      expect(state.allServices.currentPage).to.equal(1)
      expect(state.allServices.currentQuery).to.equal('test query')
      expect(state.allServices.currentStatus).to.equal('warning')
    })

    it('getAllServicesFailing', function () {
      const state = {
        allServices: {
          loading: true
        }
      }

      const error = { message: 'invalid request: page: 3, per_page: 10' }

      getAllServicesFailure(state, error)

      expect(state.allServices.loading).to.equal(false)
      expect(state.allServices.endReached).to.equal(true)
    })
  })
})
