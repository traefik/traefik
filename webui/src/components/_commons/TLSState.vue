<template>
  <q-avatar text-color="dark">
    <q-icon v-if="isTLS" name="eva-shield" />
  </q-avatar>
</template>

<script>
export default {
  name: 'TLSState',
  props: ['isTLS']
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .q-avatar{
    font-size: 32px;
    border-radius: 4px;
    .q-icon {
      color: green;
      font-size: 22px;
      margin: 0 0 0 1px;
    }
  }
</style>
