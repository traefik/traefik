<template>
  <q-header class="shadow-1">
    <section class="app-section bg-primary text-white">
      <div class="app-section-wrap app-boxed app-boxed-xl">
        <q-toolbar class="row no-wrap items-center">
          <div class="q-pr-md logo">
            <img alt="logo" src="~assets/logo.svg">
          </div>
          <q-tabs align="left" inline-label indicator-color="transparent" active-color="white" stretch>
            <q-route-tab to="/" icon="eva-home-outline" no-caps label="Dashboard" />
            <q-route-tab to="/http" icon="eva-globe-outline" no-caps label="HTTP" />
            <q-route-tab to="/tcp" icon="eva-globe-2-outline" no-caps label="TCP" />
            <q-route-tab to="/udp" icon="eva-globe-2-outline" no-caps label="UDP" />
          </q-tabs>
          <q-space />
          <q-btn @click="$q.dark.toggle()" stretch flat no-caps icon="invert_colors" :label="`${$q.dark.isActive ? 'Light' : 'Dark'} theme`" class="btn-menu" />
          <q-btn type="a" :href="`https://docs.traefik.io/${parsedVersion}`" target="_blank" stretch flat no-caps label="Documentation" class="btn-menu" />
          <q-btn v-if="version" type="a" href="https://github.com/containous/traefik/" target="_blank" stretch flat no-caps :label="`${name} ${version}`" class="btn-menu" />
        </q-toolbar>
      </div>
    </section>

    <section class="app-section text-black sub-nav" :class="{ 'bg-white': !$q.dark.isActive }">
      <div class="app-section-wrap app-boxed app-boxed-xl">
        <slot />
      </div>
    </section>
  </q-header>
</template>

<script>
import config from '../../../package'
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'NavBar',
  data () {
    return {
    }
  },
  computed: {
    ...mapGetters('core', { coreVersion: 'version' }),
    version () {
      return this.coreVersion.Version
    },
    parsedVersion () {
      if (this.version === undefined) {
        return 'master'
      }
      if (this.version === 'dev') {
        return 'master'
      }
      const matches = this.version.match(/^(v?\d+\.\d+)/)
      return matches ? 'v' + matches[1] : 'master'
    },
    name () {
      return config.productName
    }
  },
  methods: {
    ...mapActions('core', { getVersion: 'getVersion' })
  },
  created () {
    this.getVersion()
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .q-toolbar {
    min-height: 64px;
  }

  .body--dark .sub-nav {
    background-color: #0e204c;
  }

  .logo {
    display: inline-block;
    img {
      height: 24px;
    }
  }

  .q-tabs {
    color: rgba( $app-text-white, .4 );
    /deep/ .q-tabs__content {
      .q-tab__content{
        min-width: 100%;
        .q-tab__label {
          font-size: 16px;
          font-weight: 600;
        }
      }
    }
  }

  .btn-menu{
    color: rgba( $app-text-white, .4 );
    font-size: 16px;
    font-weight: 600;
  }
</style>
