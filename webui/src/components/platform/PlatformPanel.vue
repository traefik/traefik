<template>
  <side-panel :isOpen="isPlatformOpen" @onClose="closePlatform()" v-if="isOnline">
    <div class="iframe-wrapper">
      <iframe
        id="platform-iframe"
        :src="iFrameUrl"
        style="position: relative; height: 100%; width: 100%;"
        frameBorder="0"
        scrolling="yes"
        @load="onIFrameLoad"
      />
    </div>
  </side-panel>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'
import iFrameResize from 'iframe-resizer/js/iframeResizer'
import SidePanel from '../_commons/SidePanel'
import Helps from '../../_helpers/Helps'
import '../../_directives/resize'

export default {
  name: 'PlatformPanel',
  components: {
    SidePanel
  },
  async created () {
    this.getInstanceInfos()
  },
  computed: {
    ...mapGetters('platform', { isPlatformOpen: 'isOpen' }),
    ...mapGetters('core', { instanceInfos: 'version' }),
    iFrameUrl () {
      const instanceInfos = JSON.stringify(this.instanceInfos)
      const authRedirectUrl = `${window.location.href.split('?')[0]}?platform=on`
      const queryParams = `?authRedirectUrl=${encodeURIComponent(authRedirectUrl)}&instanceInfos=${encodeURIComponent(instanceInfos)}`

      return `${this.platformUrl}/instances${queryParams}`
    },
    isOnline () {
      return window.navigator.onLine
    }
  },
  methods: {
    ...mapActions('platform', { openPlatform: 'open', closePlatform: 'close' }),
    ...mapActions('core', { getInstanceInfos: 'getVersion' }),
    onIFrameLoad () {
      iFrameResize({
        log: false,
        resize: false,
        scrolling: true,
        onMessage: ({ iframe, message }) => {
          if (typeof message === 'string') {
            switch (message) {
              case 'open:profile':
                this.openPlatform()
                break
              case 'logout':
                this.closePlatform()
                break
            }
          } else if (message.type) {
            switch (message.type) {
              case 'copy-to-clipboard':
                navigator.clipboard.writeText(message.value)
                break
            }
          } else if (message.isAuthenticated) {
            this.isAuthenticated = message.isAuthenticated
          }
        }
      }, '#platform-iframe')
    }
  },
  watch: {
    $route (to, from) {
      const wasOpen = from.query && from.query.platform === 'on'
      const isOpen = to.query && to.query.platform === 'on'

      if (!wasOpen && isOpen) {
        this.openPlatform()
      }
    },
    isPlatformOpen (newValue, oldValue) {
      if (newValue !== oldValue) {
        document.querySelector('body').style.overflow = newValue ? 'hidden' : 'visible'

        this.$router.push({
          path: this.$route.path,
          query: Helps.removeEmptyObjects({
            ...this.$route.query,
            platform: newValue ? 'on' : undefined
          })
        })
      }
    }
  }
}
</script>

<style scoped>
  .iframe-wrapper {
    height: 100%;
  }
</style>
