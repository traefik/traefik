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

// ----------------------------
// Get Router By Name
// ----------------------------
export function getRouterByNameRequest (state) {
  state.routerByName.loading = true
}

export function getRouterByNameSuccess (state, body) {
  state.routerByName = { item: body, loading: false }
}

export function getRouterByNameFailure (state, error) {
  state.routerByName = { error }
}

export function getRouterByNameClear (state) {
  state.routerByName = {}
}

// ----------------------------
// Get Middleware By Name
// ----------------------------
export function getMiddlewareByNameRequest (state) {
  state.middlewareByName.loading = true
}

export function getMiddlewareByNameSuccess (state, body) {
  state.middlewareByName = { item: body, loading: false }
}

export function getMiddlewareByNameFailure (state, error) {
  state.middlewareByName = { error }
}

export function getMiddlewareByNameClear (state) {
  state.middlewareByName = {}
}
