<template>
  <q-chip
    outline
    :class="['feature-chip', {'feature-chip-on':value}, {'feature-chip-off':!value}]">
    {{ getLabel }}
  </q-chip>
</template>

<script>
export default {
  name: 'BooleanState',
  props: ['value'],
  computed: {
    getLabel () {
      return this.value ? 'ON' : 'OFF'
    }
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

    .feature-chip {
      border-radius: 12px;
      border-width: 2px;
      height: 18px;
      padding: 9px 20px;
      color: $primary;
      font-size: 16px;
      font-weight: 600;
      &-off {
        border-color: $negative;
        color: $negative;
        background-color: rgba( $negative, .1 );
      }
      &-on{
        border-color: $positive;
        color: $positive;
        background-color: rgba( $positive, .1 );
      }
    }

</style>
