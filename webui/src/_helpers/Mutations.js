import { set, get } from 'dot-prop'

export const withPagination = (type, opts = {}) => (state, data) => {
  const { isSameContext, statePath } = opts
  const currentState = get(state, statePath)

  let newState

  switch (type) {
    case 'request':
      newState = {
        loading: true
      }
      break
    case 'success':
      const { body, page } = data
      newState = {
        ...currentState,
        items: [
          ...(isSameContext && currentState.items && page !== 1 ? currentState.items : []),
          ...(body.data || [])
        ],
        currentPage: page,
        total: body.total,
        loading: false
      }
      break
    case 'failure':
      newState = {
        loading: false,
        error: data,
        endReached: data.message.includes('invalid request: page:')
      }
      break
  }

  if (newState) {
    set(state, statePath, newState)
  }
}
