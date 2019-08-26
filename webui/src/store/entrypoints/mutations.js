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
