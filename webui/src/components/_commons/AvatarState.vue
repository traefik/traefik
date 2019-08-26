<template>
  <q-avatar :color="state" text-color="white">
    <q-icon v-if="state === 'positive'" name="eva-checkmark-circle-2" />
    <q-icon v-if="state === 'warning'" name="eva-alert-circle" />
    <q-icon v-if="state === 'negative'" name="eva-alert-triangle" />
  </q-avatar>
</template>

<script>
export default {
  name: 'AvatarState',
  props: ['state']
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .q-avatar{
    font-size: 32px;
    border-radius: 4px;
    .q-icon {
      font-size: 22px;
      margin: 0 0 0 1px;
    }
  }
</style>
