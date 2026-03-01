import { setupServer } from 'msw/node'

import { getHandlers } from './handlers'

export const server = setupServer(...getHandlers(true))
