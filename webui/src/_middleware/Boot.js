import { APP } from '../_helpers/APP'
import Helps from '../_helpers/Helps'

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
        platformUrl () {
          return APP.config.platformUrl
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
        capFirstLetter (value) {
          return Helps.capFirstLetter(value)
        }
      },
      created () {
      }
    })
  }
}

export default Boot
