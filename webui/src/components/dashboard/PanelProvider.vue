<template>
  <q-card flat bordered>
    <q-card-section>
      <div class="row items-center no-wrap">
        <div class="col text-center">
          <q-avatar class="provider-logo">
            <q-icon :name="`img:statics/providers/${getNameLogo}.svg`" />
          </q-avatar>
        </div>
      </div>
    </q-card-section>
    <q-card-section>
      <div class="text-h6 text-center text-weight-bold">
        {{getName}}
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
export default {
  name: 'PanelProvider',
  props: ['name'],
  computed: {
    getName () {
      return this.name
    },
    getNameLogo () {
      return this.getName.toLowerCase()
    }
  }
}
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
