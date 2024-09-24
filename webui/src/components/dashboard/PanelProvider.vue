<template>
  <q-card
    flat
    bordered
  >
    <q-card-section>
      <div class="row items-center no-wrap">
        <div class="col text-center">
          <q-avatar
            class="provider-logo"
            font-size="inherit"
          >
            <q-icon :name="`img:${getLogoPath}`" />
          </q-avatar>
        </div>
      </div>
    </q-card-section>
    <q-card-section>
      <div class="text-h6 text-center text-weight-bold">
        {{ getName }}
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'PanelProvider',
  props: {
    name: {
      type: String,
      default: '',
      required: false
    }
  },
  computed: {
    getName () {
      return this.name
    },
    getLogoPath () {
      const name = this.getName.toLowerCase()

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
  @import "../../css/sass/variables";

  .provider-logo {
    width: 52px;
    height: 52px;
    img {
      width: 100%;
      height: 100%;
    }
  }
</style>
