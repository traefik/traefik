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
          v-bind:class="['feature-chip', {'feature-chip-string':isString}, {'feature-chip-boolean':isBoolean}, {'feature-chip-boolean-true':isTrue}]">
          {{getVal}}
        </q-chip>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
export default {
  name: 'PanelFeature',
  props: ['featureKey', 'featureVal'],
  computed: {
    isString () {
      return this.$_.isString(this.featureVal)
    },
    isBoolean () {
      return this.$_.isBoolean(this.featureVal) || this.featureVal === ''
    },
    isTrue () {
      return this.isBoolean && this.featureVal === true
    },
    getVal () {
      if (this.featureVal === true) {
        return 'ON'
      } else if (this.featureVal === false || this.featureVal === '') {
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
    color: $primary;
    &-string{
      border-color: $app-text-grey;
      font-size: 24px;
      color: $app-text-grey;
      background-color: rgba( $app-text-grey, .1 );
    }
    &-boolean{
      font-size: 40px;
      font-weight: 700;
      border-color: $negative;
      color: $negative;
      background-color: rgba( $negative, .1 );
      &-true{
        border-color: $positive;
        color: $positive;
        background-color: rgba( $positive, .1 );
      }
    }
  }
</style>
