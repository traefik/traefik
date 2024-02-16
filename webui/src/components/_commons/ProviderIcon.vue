<template>
  <q-avatar class="provider-logo">
    <q-icon :name="`img:${getLogoPath}`" />
  </q-avatar>
</template>

<script>
import { defineComponent } from 'vue'

export default defineComponent({
  props: {
    name: String
  },
  computed: {
    getLogoPath () {
      const name = this.name.toLowerCase()

      if (name.startsWith('plugin-')) {
        return 'providers/plugin.svg'
      }
      if (name.startsWith('consul-')) {
        return 'providers/consul.svg'
      }
      if (name.startsWith('consulcatalog-')) {
        return 'providers/consulcatalog.svg'
      }
      if (name.startsWith('nomad-')) {
        return 'providers/nomad.svg'
      }

      return `providers/${name}.svg`
    }
  }
})
</script>

<style scoped lang="scss">
  .provider-logo {
    width: 32px;
    height: 32px;
    img {
      width: 100%;
      height: 100%;
    }
  }
</style>
