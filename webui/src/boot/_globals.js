import { APP } from '../_helpers/APP'
import Boot from '../_middleware/Boot'

export default async ({ app, router, store }) => {
  app.use(Boot)

  APP.root = app
  APP.router = router
  APP.store = store
}
