<template>
  <transition name="slide" v-if="!instanceIsRegistered && isOnline && !platformNotificationIsHidden">
    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-md">
        <div class="platform-notification q-pl-lg q-pr-lg q-pt-md q-pb-md">
          <p>
            <q-avatar color="accent" text-color="white" class="icon">
              <q-icon name="eva-alert-circle" />
            </q-avatar>
            This Traefik Instance is not registered in your Traefik Pilot account yet.
          </p>
          <platform-action-button label="Register Traefik instance" @click="openPlatform" />
          <q-btn round size="xs" color="accent" icon="close" class="btn close-btn" @click="hidePlatformNotification"></q-btn>
        </div>
      </div>
    </section>
  </transition>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'
import PlatformActionButton from './PlatformActionButton'

export default {
  name: 'PlatformNotification',
  components: { PlatformActionButton },
  created () {
    this.getInstanceInfos()
    console.log(`localStorage.getItem('platform_notification-is-hidden')`, localStorage.getItem('platform_notification-is-hidden'))
    try {
      if (localStorage.getItem('platform_notification-is-hidden') === 'true') {
        this.hidePlatformNotification()
      }
    } catch (e) {
      console.error('error reading localStorage', e)
    }
  },
  computed: {
    ...mapGetters('platform', { isPlatformOpen: 'isOpen', platformNotificationIsHidden: 'notificationIsHidden' }),
    ...mapGetters('core', { instanceInfos: 'version' }),
    instanceIsRegistered () {
      return !!(this.instanceInfos && this.instanceInfos.uuid && this.instanceInfos.uuid.length > 0)
    },
    isOnline () {
      return window.navigator.onLine
    }
  },
  methods: {
    ...mapActions('platform', { openPlatform: 'open', hidePlatformNotification: 'hideNotification' }),
    ...mapActions('core', { getInstanceInfos: 'getVersion' })
  },
  watch: {
    isPlatformOpen (newValue, oldValue) {
      const isClosed = !newValue && oldValue
      if (isClosed) {
        this.getInstanceInfos()
      }
    },
    platformNotificationIsHidden (newValue, oldValue) {
      if (newValue !== oldValue) {
        try {
          localStorage.setItem('platform_notification-is-hidden', newValue ? 'true' : 'false')
        } catch (e) {
          console.error('error writing localStorage', e)
        }
      }
    }
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .platform-notification {
    position: relative;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-radius: 6px;
    color: $accent;
    background-color: rgba($accent, 0.1);
    font-size: 16px;
  }

  .close-btn {
    position: absolute;
    top: -8px;
    right: -8px;
  }

  p {
    margin: 0;
  }

  .icon {
    width: 32px;
    height: 32px;
    margin-right: 20px;
    border-radius: 4px;
  }

  .slide-enter-active,
  .slide-leave-active {
    transition: transform 0.5s ease;
  }

  .slide-enter,
  .slide-leave-to {
    transform: translateX(100%);
    transition: all 150ms ease-in 0s;
  }
</style>
