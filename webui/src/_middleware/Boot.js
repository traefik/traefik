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
      },
      filters: {
        middlewareTypeLabel (value) {
          // TODO - add all types to middlewares
          // fake function, remplace for optimized function
          let label = value
          if (value === 'redirectscheme') {
            label = 'redirectScheme'
          }
          if (value === 'basicauth') {
            label = 'basicAuth'
          }
          return label
        }
      },
      created () {
      }
    })
  }
}

export default Boot
