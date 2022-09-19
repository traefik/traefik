<template>
  <q-avatar class="provider-logo">
    <q-icon :name="`img:${getLogoPath}`" />
  </q-avatar>
</template>

<script>
export default {
  props: ['name'],
  computed: {
    getLogoPath () {
      const name = this.name.toLowerCase()

      if (name.startsWith('plugin-')) {
        return 'statics/providers/plugin.svg'
      }
      if (name.startsWith('consul-')) {
        return `statics/providers/consul.svg`
      }
      if (name.startsWith('consulcatalog-')) {
        return `statics/providers/consulcatalog.svg`
      }
      if (name.startsWith('nomad-')) {
        return `statics/providers/nomad.svg`
      }

      return `statics/providers/${name}.svg`
    }
  }
}
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
