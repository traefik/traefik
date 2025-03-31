// ----------------------------
// Get Overview
// ----------------------------
export function getOverviewRequest (state) {
  state.allOverview.loading = true
}

export function getOverviewSuccess (state, body) {
  state.allOverview = { items: body, loading: false }
}

export function getOverviewFailure (state, error) {
  state.allOverview = { error }
}

export function getOverviewClear (state) {
  state.allOverview = {}
}

// ----------------------------
// Get Version
// ----------------------------
export function getVersionSuccess (state, body) {
  state.version = body
  state.version.disableDashboardAd = !!body.disableDashboardAd // Ensures state.version.disableDashboardAd is defined
}
