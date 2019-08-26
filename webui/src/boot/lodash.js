import lodash from 'lodash'

import { APP } from '../_helpers/APP'

export default async ({ app, Vue }) => {
  Vue.prototype.$_ = app._ = APP._ = lodash
}
