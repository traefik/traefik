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
// Get All Services
// ----------------------------
export function getAllServicesRequest (state) {
  state.allServices.loading = true
}

export function getAllServicesSuccess (state, body) {
  state.allServices = { items: body.data, total: body.total, loading: false }
}

export function getAllServicesFailure (state, error) {
  state.allServices = { error }
}

export function getAllServicesClear (state) {
  state.allServices = {}
}

// ----------------------------
// Get Service By Name
// ----------------------------
export function getServiceByNameRequest (state) {
  state.serviceByName.loading = true
}

export function getServiceByNameSuccess (state, body) {
  state.serviceByName = { item: body, loading: false }
}

export function getServiceByNameFailure (state, error) {
  state.serviceByName = { error }
}

export function getServiceByNameClear (state) {
  state.serviceByName = {}
}
