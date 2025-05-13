<template>
  <div id="q-app">
    <router-view />
  </div>
</template>

<script>
import { APP } from './_helpers/APP'
import { mapGetters } from 'vuex'

export default {
  name: 'App',
  computed: {
    ...mapGetters('core', { coreVersion: 'version' })
  },
  watch: {
    '$q.dark.mode' (val) {
      if (val !== null) {
        localStorage.setItem('traefik-dark', val)
      }
    }
  },
  beforeCreate () {
    // Set vue instance
    APP.vue = () => this.$root

    // debug
    console.log('Quasar -> ', this.$q.version)

    // Get stored theme or default to 'auto'
    const storedTheme = localStorage.getItem('traefik-dark')
    if (storedTheme === 'true') {
      this.$q.dark.set(true)
    } else if (storedTheme === 'false') {
      this.$q.dark.set(false)
    } else {
      this.$q.dark.set('auto')
    }
  }
}
</script>

<style>
</style>
