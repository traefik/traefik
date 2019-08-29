// ----------------------------
// Get All Routers
// ----------------------------
export function getAllRoutersRequest (state) {
  state.allRouters.loading = true
}

export function getAllRoutersSuccess (state, body) {
  state.allRouters = { items: body.data, total: body.total, loading: false }
}

export function getAllRoutersFailure (state, error) {
  state.allRouters = { error }
}

export function getAllRoutersClear (state) {
  state.allRouters = {}
}
