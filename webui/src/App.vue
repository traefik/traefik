<template>
  <div id="q-app">
    <router-view />
    <platform-panel />
  </div>
</template>

<script>
import { APP } from './_helpers/APP'
import PlatformPanel from './components/platform/PlatformPanel'

export default {
  name: 'App',
  components: {
    PlatformPanel
  },
  beforeCreate () {
    // Set vue instance
    APP.vue = () => this.$root

    // debug
    console.log('Quasar -> ', this.$q.version)

    this.$q.dark.set(localStorage.getItem('traefik-dark') === 'true')
  },
  watch: {
    '$q.dark.isActive' (val) {
      localStorage.setItem('traefik-dark', val)
    }
  }
}
</script>

<style>
</style>
