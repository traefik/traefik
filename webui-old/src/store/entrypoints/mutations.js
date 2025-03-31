// ----------------------------
// Get All
// ----------------------------
export function getAllRequest (state) {
  state.all.loading = true
}

export function getAllSuccess (state, body) {
  state.all = { items: body, loading: false }
}

export function getAllFailure (state, error) {
  state.all = { error }
}

export function getAllClear (state) {
  state.all = {}
}

// ----------------------------
// Get By Name
// ----------------------------
export function getByNameRequest (state) {
  state.byName.loading = true
}

export function getByNameSuccess (state, body) {
  state.byName = { item: body, loading: false }
}

export function getByNameFailure (state, error) {
  state.byName = { error }
}

export function getByNameClear (state) {
  state.byName = {}
}
