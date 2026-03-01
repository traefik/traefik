import { http, passthrough } from 'msw'

import apiEntrypoints from './data/api-entrypoints.json'
import apiHttpMiddlewares from './data/api-http_middlewares.json'
import apiHttpRouters from './data/api-http_routers.json'
import apiHttpServices from './data/api-http_services.json'
import apiOverview from './data/api-overview.json'
import apiTcpMiddlewares from './data/api-tcp_middlewares.json'
import apiTcpRouters from './data/api-tcp_routers.json'
import apiTcpServices from './data/api-tcp_services.json'
import apiUdpRouters from './data/api-udp_routers.json'
import apiUdpServices from './data/api-udp_services.json'
import apiVersion from './data/api-version.json'
import eeApiErrors from './data/ee-api-errors.json'
import { listHandlers } from './utils'

export const getHandlers = (noDelay: boolean = false) => [
  ...listHandlers('/api/entrypoints', apiEntrypoints, noDelay, true),
  ...listHandlers('/api/errors', eeApiErrors, noDelay),
  ...listHandlers('/api/http/middlewares', apiHttpMiddlewares, noDelay),
  ...listHandlers('/api/http/routers', apiHttpRouters, noDelay),
  ...listHandlers('/api/http/services', apiHttpServices, noDelay),
  ...listHandlers('/api/overview', apiOverview, noDelay),
  ...listHandlers('/api/tcp/middlewares', apiTcpMiddlewares, noDelay),
  ...listHandlers('/api/tcp/routers', apiTcpRouters, noDelay),
  ...listHandlers('/api/tcp/services', apiTcpServices, noDelay),
  ...listHandlers('/api/udp/routers', apiUdpRouters, noDelay),
  ...listHandlers('/api/udp/services', apiUdpServices, noDelay),
  ...listHandlers('/api/version', apiVersion, noDelay),
  http.get('*.tsx', () => passthrough()),
  http.get('/img/*', () => passthrough()),
]
