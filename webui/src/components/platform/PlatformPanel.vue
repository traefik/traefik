<template>
  <side-panel
    :isOpen="isPlatformOpen"
    @onClose="closePlatform()"
    v-if="isOnline"
  >
    <div class="iframe-wrapper">
      <iframe
        id="platform-iframe"
        :src="iFrameUrl"
        v-resize="resizeOpts"
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
import qs from 'query-string'
import SidePanel from '../_commons/SidePanel'
import Helps from '../../_helpers/Helps'
import '../../_directives/resize'

export default {
  name: 'PlatformPanel',
  components: {
    SidePanel
  },
  data () {
    return {
      resizeOpts: {
        log: false,
        resize: false,
        scrolling: true,
        onMessage: ({ iframe, message }) => {
          if (typeof message === 'string') {
            // 1st condition for backward compatibility
            if (message === 'open:profile') {
              this.openPlatform('/')
            } else if (message.includes('open:')) {
              this.openPlatform(message.split('open:')[1])
            } else if (message === 'logout') {
              this.closePlatform()
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
      }
    }
  },
  created () {
    this.getInstanceInfos()
  },
  computed: {
    ...mapGetters('platform', { isPlatformOpen: 'isOpen', platformPath: 'path' }),
    ...mapGetters('core', { instanceInfos: 'version' }),
    iFrameUrl () {
      const instanceInfos = JSON.stringify(this.instanceInfos)
      const authRedirectUrl = `${window.location.href.split('?')[0]}?platform=${this.platformPath}`

      return qs.stringifyUrl({ url: `${this.platformUrl}${this.platformPath}`, query: { authRedirectUrl, instanceInfos } })
    },
    isOnline () {
      return window.navigator.onLine
    }
  },
  methods: {
    ...mapActions('platform', { openPlatform: 'open', closePlatform: 'close' }),
    ...mapActions('core', { getInstanceInfos: 'getVersion' })
  },
  watch: {
    $route (to, from) {
      const wasOpen = from.query && from.query.platform
      const isOpen = to.query && to.query.platform

      if (!wasOpen && isOpen) {
        this.openPlatform(to.query.platform)
      }
    },
    isPlatformOpen (newValue, oldValue) {
      if (newValue !== oldValue) {
        document.querySelector('body').style.overflow = newValue ? 'hidden' : 'visible'

        this.$router.push({
          path: this.$route.path,
          query: Helps.removeEmptyObjects({
            ...this.$route.query,
            platform: this.platformPath ? this.platformPath : undefined
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
