<template>
  <q-layout view="hHh lpR fFf">

    <nav-bar>
      <router-view name="NavBar" />
    </nav-bar>

    <q-page-container>
      <router-view />
    </q-page-container>

  </q-layout>
</template>

<script>
import NavBar from '../components/_commons/NavBar'

export default {
  name: 'Default',
  components: {
    NavBar
  },
  data () {
    return {
    }
  }
}
</script>

<style scoped lang="scss">

</style>
