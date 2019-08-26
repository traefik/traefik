import { APP } from '../_helpers/APP'

const Boot = {
  install (Vue, options) {
    Vue.mixin({
      data () {
        return {
        }
      },
      computed: {
        api () {
          return APP.config.apiUrl
        },
        env () {
          return APP.config.env
        }
      },
      methods: {
      },
      filters: {
      },
      created () {
      }
    })
  }
}

export default Boot
