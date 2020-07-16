import Vue from 'vue'
import iFrameResize from 'iframe-resizer/js/iframeResizer'

Vue.directive('resize', {
  bind: function (el, { value = {} }) {
    el.addEventListener('load', () => iFrameResize(value, el))
  }
})
