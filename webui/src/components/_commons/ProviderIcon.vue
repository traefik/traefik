<template>
  <q-avatar class="provider-logo">
    <q-icon :name="`img:statics/providers/${name}.svg`" />
  </q-avatar>
</template>

<script>
export default {
  props: ['name']
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
