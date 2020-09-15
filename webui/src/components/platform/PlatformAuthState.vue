<template>
  <div class="iframe-wrapper" v-if="isOnline">
    <iframe
      id="platform-auth-state"
      :src="iFrameUrl"
      v-if="renderIrame"
      v-resize="resizeOpts"
      height="64px"
      frameBorder="0"
    />
  </div>

</template>

<script>
import { mapActions, mapGetters } from 'vuex'
import qs from 'query-string'
import '../../_directives/resize'

export default {
  name: 'PlatformPanel',
  data () {
    return {
      renderIrame: true,
      resizeOpts: {
        log: false,
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
    isOnline () {
      return window.navigator.onLine
    },
    iFrameUrl () {
      const instanceInfos = JSON.stringify(this.instanceInfos)
      const authRedirectUrl = `${window.location.href.split('?')[0]}?platform=${this.platformPath}`

      return qs.stringifyUrl({ url: `${this.platformUrl}/partials/auth-state`, query: { authRedirectUrl, instanceInfos } })
    }
  },
  methods: {
    ...mapActions('platform', { openPlatform: 'open' }, { closePlatform: 'close' }),
    ...mapActions('core', { getInstanceInfos: 'getVersion' })
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

<style scoped lang="scss">
  @import "../../css/sass/variables";

  #platform-auth-state {
    width: 1px;
    min-width: 296px;
  }
</style>
