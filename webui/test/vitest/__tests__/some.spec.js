import { describe, expect, it } from 'vitest'
import store from 'src/store/http/index.js'

const {
  getAllRoutersRequest
} = store.mutations

describe('http mutations', function () {
  /* Routers */
  describe('http routers mutations', function () {
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
  })
})
