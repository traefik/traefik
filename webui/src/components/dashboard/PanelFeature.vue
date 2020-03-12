<template>
  <q-card flat bordered v-bind:class="['panel-feature']">
    <q-card-section>
      <div class="row items-center no-wrap">
        <div class="col">
          <div class="text-subtitle2">{{featureKey}}</div>
        </div>
      </div>
    </q-card-section>
    <q-card-section>
      <div class="text-h3 text-center text-weight-bold">
        <q-chip
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
      return typeof this.featureVal === 'string'
    },
    isBoolean () {
      return typeof this.featureVal === 'boolean' || this.featureVal === ''
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

  .panel-feature {
    .text-subtitle2 {
      font-weight: 600;
      letter-spacing: 3px;
      color: $app-text-grey;
      text-transform: uppercase;
      text-align: center;
    }
  }

  .feature-chip {
    border-radius: 12px;
    border-width: 2px;
    height: 56px;
    padding: 12px 24px;
    color: $primary;
    &-string{
      border-color: $app-text-grey;
      font-size: 20px;
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

  .body--dark {
    .feature-chip-string {
      background-color: rgba( $app-text-grey, .3 );
    }
    .feature-chip-boolean {
      background-color: rgba( $negative, .3 );
      &-true {
        background-color: rgba( $positive, .3 );
      }
    }
  }
</style>
