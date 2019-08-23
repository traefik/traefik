<template>
  <q-card flat bordered>
    <q-card-section>
      <div class="row items-center no-wrap">
        <div class="col">
          <div class="text-subtitle2 text-uppercase text-center text-app-grey" style="letter-spacing: 3px;">{{featureKey}}</div>
        </div>
      </div>
    </q-card-section>
    <q-card-section>
      <div class="text-h3 text-center text-weight-bold">
        <q-chip
          outline
          color="primary"
          text-color="white"
          v-bind:class="['feature-chip', {'feature-chip-string':isString()}, {'feature-chip-boolean':isBoolean()}, {'feature-chip-boolean-true':isTrue()}]">
          {{getVal()}}
        </q-chip>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
export default {
  name: 'PanelFeature',
  props: ['featureKey', 'featureVal'],
  methods: {
    isString () {
      return this.$_.isString(this.featureVal)
    },
    isBoolean () {
      return this.$_.isBoolean(this.featureVal)
    },
    isTrue () {
      return this.isBoolean() && this.featureVal === true
    },
    getVal () {
      if (this.featureVal === true) {
        return 'ON'
      } else if (this.featureVal === false) {
        return 'OFF'
      } else {
        return this.featureVal
      }
    }
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .feature-chip {
    border-radius: 12px;
    border-width: 2px;
    height: 56px;
    padding: 12px 24px;
    &-string{
      border-color: $app-text-grey;
      font-size: 24px;
      color: $app-text-grey !important;
      background-color: rgba( $app-text-grey, .1 );
    }
    &-boolean{
      font-size: 40px;
      font-weight: 700;
      border-color: $negative;
      color: $negative !important;
      background-color: rgba( $negative, .1 );
      &-true{
        border-color: $positive;
        color: $positive !important;
        background-color: rgba( $positive, .1 );
      }
    }
  }
</style>
