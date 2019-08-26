<template>
  <div id="q-app">
    <router-view />
  </div>
</template>

<script>
import { APP } from './_helpers/APP'

export default {
  name: 'App',
  beforeCreate () {
    // Set vue instance
    APP.vue = () => this.$root

    // debug
    console.log('Quasar -> ', this.$q.version)
  }
}
</script>

<style>
</style>
