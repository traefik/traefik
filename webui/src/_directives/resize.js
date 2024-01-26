import iFrameResize from 'iframe-resizer/js/iframeResizer'
import { APP } from './_helpers/APP'

APP.vue.directive('resize', {
  bind: function (el, { value = {} }) {
    el.addEventListener('load', () => iFrameResize(value, el))
  }
})
