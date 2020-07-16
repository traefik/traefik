<template>
  <div class="iframe-wrapper" v-if="isOnline">
    <iframe
      v-if="renderIrame"
      id="platform-auth-state"
      :src="AuthStateIFrameUrl"
      @load="onAuthStateIFrameLoad"
      height="64px"
      frameBorder="0"
    />
  </div>

</template>

<script>
import { mapActions, mapGetters } from 'vuex'
import iFrameResize from 'iframe-resizer/js/iframeResizer'
import '../../_directives/resize'

export default {
  name: 'PlatformPanel',
  data () {
    return {
      renderIrame: true
    }
  },
  computed: {
    ...mapGetters('platform', { isPlatformOpen: 'isOpen' }),
    isOnline () {
      return window.navigator.onLine
    }
  },
  methods: {
    ...mapActions('platform', { openPlatform: 'open' }, { closePlatform: 'close' }),
    onAuthStateIFrameLoad () {
      iFrameResize({
        log: false,
        resize: true,
        onMessage: ({ iframe, message }) => {
          if (typeof message === 'string') {
            if (message === 'open:profile') {
              this.openPlatform()
            }

            if (message === 'logout') {
              this.closePlatform()
            }
          }
        }
      }, '#platform-auth-state')
    }
  },
  created () {
    const authRedirectUrl = `${window.location.href.split('?')[0]}?platform=on`
    const queryParams = `?authRedirectUrl=${encodeURIComponent(authRedirectUrl)}`

    this.AuthStateIFrameUrl = `${this.platformUrl}/partials/auth-state${queryParams}`
  },
  watch: {
    isPlatformOpen (isOpen, wasOpen) {
      if (!isOpen && wasOpen) {
        this.renderIrame = false
        this.$nextTick().then(() => {
          this.renderIrame = true
        })
      }
    }
  }
}
</script>
