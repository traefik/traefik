import { describe, expect, it } from 'vitest'
import store from './index.js'

const {
  getAllRoutersRequest,
  getAllRoutersSuccess,
  getAllRoutersFailure,
  getAllServicesRequest,
  getAllServicesSuccess,
  getAllServicesFailure
} = store.mutations

describe('udp mutations', function () {
  /* Routers */
  describe('udp routers mutations', function () {
    it('getAllRoutersRequest', function () {
      const state = {
        allRouters: {
          items: [{}, {}, {}]
        }
      }

      getAllRoutersRequest(state)

      expect(state.allRouters.loading).to.equal(true)
      expect(state.allRouters.items.length).to.equal(3)
    })

    it('getAllRoutersSuccess page 1', function () {
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

    it('getAllRoutersSuccess page 2', function () {
      const state = {
        allRouters: {
          loading: false,
          items: [{ id: 1 }, { id: 2 }, { id: 3 }],
          total: 3,
          currentPage: 1,
          currentQuery: 'test query',
          currentStatus: 'warning'
        }
      }

      const data = {
        body: {
          data: [{ id: 4 }, { id: 5 }, { id: 6 }, { id: 7 }],
          total: 4
        },
        query: 'test query',
        status: 'warning',
        page: 2
      }

      getAllRoutersSuccess(state, data)

      expect(state.allRouters.loading).to.equal(false)
      expect(state.allRouters.total).to.equal(7)
      expect(state.allRouters.items.length).to.equal(7)
      expect(state.allRouters.currentPage).to.equal(2)
      expect(state.allRouters.currentQuery).to.equal('test query')
      expect(state.allRouters.currentStatus).to.equal('warning')
    })

    it('getAllRoutersFailing', function () {
      const state = {
        allRouters: {
          items: [{}, {}, {}],
          loading: true
        }
      }

      const error = { message: 'invalid request: page: 3, per_page: 10' }

      getAllRoutersFailure(state, error)

      expect(state.allRouters.loading).to.equal(false)
      expect(state.allRouters.endReached).to.equal(true)
      expect(state.allRouters.items.length).to.equal(3)
    })
  })

  /* Services */
  describe('udp services mutations', function () {
    it('getAllServicesRequest', function () {
      const state = {
        allServices: {
          items: [{}, {}, {}]
        }
      }

      getAllServicesRequest(state)

      expect(state.allServices.loading).to.equal(true)
      expect(state.allServices.items.length).to.equal(3)
    })

    it('getAllServicesSuccess page 1', function () {
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

    it('getAllServicesSuccess page 2', function () {
      const state = {
        allServices: {
          loading: false,
          items: [{ id: 1 }, { id: 2 }, { id: 3 }],
          total: 3,
          currentPage: 1,
          currentQuery: 'test query',
          currentStatus: 'warning'
        }
      }

      const data = {
        body: {
          data: [{ id: 4 }, { id: 5 }, { id: 6 }, { id: 7 }],
          total: 4
        },
        query: 'test query',
        status: 'warning',
        page: 2
      }

      getAllServicesSuccess(state, data)

      expect(state.allServices.loading).to.equal(false)
      expect(state.allServices.total).to.equal(7)
      expect(state.allServices.items.length).to.equal(7)
      expect(state.allServices.currentPage).to.equal(2)
      expect(state.allServices.currentQuery).to.equal('test query')
      expect(state.allServices.currentStatus).to.equal('warning')
    })

    it('getAllServicesFailing', function () {
      const state = {
        allServices: {
          items: [{}, {}, {}],
          loading: true
        }
      }

      const error = { message: 'invalid request: page: 3, per_page: 10' }

      getAllServicesFailure(state, error)

      expect(state.allServices.loading).to.equal(false)
      expect(state.allServices.endReached).to.equal(true)
      expect(state.allServices.items.length).to.equal(3)
    })
  })
})
