<template>
  <div id="q-app">
    <router-view />
  </div>
</template>

<script>
import { APP } from './_helpers/APP'
import { mapGetters } from 'vuex'
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'App',
  computed: {
    ...mapGetters('core', { coreVersion: 'version' })
  },
  watch: {
    '$q.dark.isActive' (val) {
      localStorage.setItem('traefik-dark', val)
    }
  },
  beforeCreate () {
    // Set vue instance
    APP.vue = () => this.$root

    // debug
    console.log('Quasar -> ', this.$q.version)

    this.$q.dark.set(localStorage.getItem('traefik-dark') === 'true')
  }
})
</script>
