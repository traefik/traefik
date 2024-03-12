import iframeResize from 'iframe-resizer/js/iframeResizer'

const resize = {
  mounted (el, binding) {
    const options = binding.value || {}
    el.addEventListener('load', () => iframeResize(options, el))
  },
  unmounted (el) {
    const resizableEl = el
    if (resizableEl.iFrameResizer) {
      resizableEl.iFrameResizer.removeListeners()
    }
  }
}

export default resize
