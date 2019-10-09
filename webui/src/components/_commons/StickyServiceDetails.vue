<template>
    <div>
      <q-card-section>
        <div class="row items-start no-wrap">
          <div class="text-subtitle1">Sticky: Cookie </div>
        </div>
      </q-card-section>
      <q-card-section>
        <div class="row items-start no-wrap">
          <div class="col" v-if="sticky.cookie && sticky.cookie.name">
            <div class="text-subtitle2">NAME</div>
            <q-chip
              dense
              class="app-chip app-chip-entry-points">
              {{ sticky.cookie.name }}
            </q-chip>
          </div>
        </div>
      </q-card-section>
      <q-card-section>
        <div class="row items-start no-wrap">
          <div class="col">
            <div class="text-subtitle2">SECURE</div>
            <boolean-state :value="sticky.cookie.secure"/>
          </div>

          <div class="col">
            <div class="text-subtitle2">HTTP Only</div>
            <boolean-state :value="sticky.cookie.httpOnly"/>
          </div>
        </div>
      </q-card-section>
    </div>
</template>

<script>
import BooleanState from './BooleanState'

export default {
  name: 'StickyServiceDetails',
  components: {
    BooleanState
  },
  props: ['sticky', 'dense']
}
</script>
<style scoped lang="scss">
  @import "../../css/sass/variables";

  .q-card__section {
    padding: 24px;
    + .q-card__section {
      padding-top: 0;
    }
  }

  .text-subtitle2 {
    font-size: 11px;
    color: $app-text-grey;
    line-height: 16px;
    margin-bottom: 4px;
    text-align: left;
    letter-spacing: 2px;
    font-weight: 600;
    text-transform: uppercase;
  }

</style>
