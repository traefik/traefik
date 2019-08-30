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
        },
        appThumbStyle () {
          return {
            right: '2px',
            borderRadius: '2px',
            backgroundColor: '#dcdcdc',
            width: '6px',
            opacity: 0.75
          }
        }
      },
      methods: {
        middlewareLabel (item) {
          // TODO - add all types to middlewares
          // fake function
          let label = ''
          if (item.redirectScheme) {
            label = 'redirectScheme'
          }
          if (item.basicAuth) {
            label = 'basicAuth'
          }
          return label
        }
      },
      filters: {
      },
      created () {
      }
    })
  }
}

export default Boot
