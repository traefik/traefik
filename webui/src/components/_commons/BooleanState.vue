<template>
  <div class="block-right-text">
    <q-avatar :color="value ? 'positive' : 'negative'" text-color="white">
      <q-icon v-if="value" name="eva-toggle-right" />
      <q-icon v-if="!value" name="eva-toggle-left" />
    </q-avatar>
    <div v-bind:class="['block-right-text-label', `block-right-text-label-${!!value}`]">{{value ? 'True' : 'False'}}</div>
  </div>
</template>

<script>
export default {
  name: 'BooleanState',
  props: ['value']
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

  .block-right-text{
    height: 32px;
    line-height: 32px;
    .q-avatar{
      float: left;
    }
    &-label{
      font-size: 14px;
      font-weight: 600;
      color: $app-text-grey;
      float: left;
      margin-left: 10px;
      text-transform: capitalize;
      &-true {
        color: $positive;
      }
      &-false {
        color: $negative;
      }
    }
  }

</style>
