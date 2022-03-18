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
