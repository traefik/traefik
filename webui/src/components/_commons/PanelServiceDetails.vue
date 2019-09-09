<template>
  <q-card flat bordered v-bind:class="['panel-service-details', {'panel-service-details-dense':isDense}]">
    <q-scroll-area :thumb-style="appThumbStyle" style="height:100%;">
      <q-card-section>
        <div class="row items-start no-wrap">
          <div class="col">
            <div class="text-subtitle2">STATUS</div>
            <div class="block-right-text">
              <avatar-state :state="data.status | status "/>
              <div v-bind:class="['block-right-text-label', `block-right-text-label-${data.status}`]">{{data.status | statusLabel}}</div>
            </div>
          </div>
          <div class="col">
            <div class="text-subtitle2">PROVIDER</div>
            <div class="block-right-text">
              <q-avatar class="provider-logo">
                <q-icon :name="`img:statics/providers/${data.provider}.svg`" />
              </q-avatar>
              <div class="block-right-text-label">{{data.provider}}</div>
            </div>
          </div>
        </div>
      </q-card-section>
      <q-card-section v-if="data.type || data.strategy">
        <div class="row items-start no-wrap">
          <div class="col" v-if="data.type">
            <div class="text-subtitle2">TYPE</div>
            <q-chip
              outline
              dense
              class="app-chip app-chip-entry-points">
              {{ data.type }}
            </q-chip>
          </div>
          <div class="col" v-if="data.mirroring && data.mirroring.service">
            <div class="text-subtitle2">Main Service</div>
            <q-chip
              outline
              dense
              class="app-chip app-chip-name">
              {{ data.mirroring.service }}
            </q-chip>
          </div>
          <div class="col" v-if="data.strategy">
            <div class="text-subtitle2">STRATEGY</div>
            <q-chip
              outline
              dense
              class="app-chip app-chip-name">
              {{ data.strategy }}
            </q-chip>
          </div>
        </div>
      </q-card-section>
      <q-separator v-if="data.weighted" />
      <q-card-section v-if="data.weighted" >
        <div class="row items-start no-wrap">
          <div class="text-subtitle1">Sticky: Cookie</div>
        </div>
        <br/>
        <div class="row items-start no-wrap">
          <div class="col" v-if="data.type">
            <div class="text-subtitle2">NAME</div>
            <q-chip
              outline
              dense
              class="app-chip app-chip-entry-points">
              {{ data.weighted.sticky.cookie.name }}
            </q-chip>
          </div>
          </div>
      </q-card-section>
      <q-card-section v-if="data.weighted" >
        <div class="row items-start no-wrap">
          <div class="col">
            <div class="text-subtitle2">SECURE</div>
            <boolean-state :value="data.weighted.sticky.cookie.secure"/>
          </div>

          <div class="col">
            <div class="text-subtitle2">HTTP Only</div>
            <boolean-state :value="data.weighted.sticky.cookie.httpOnly"/>
          </div>
        </div>
      </q-card-section>
    </q-scroll-area>
  </q-card>
</template>

<script>
import AvatarState from './AvatarState'
import BooleanState from './BooleanState'

export default {
  name: 'PanelServiceDetails',
  props: ['data', 'dense'],
  components: {
    BooleanState,
    AvatarState
  },
  computed: {
    isDense () {
      return this.dense !== undefined
    }
  },
  filters: {
    status (value) {
      let status = value
      if (value === 'enabled') {
        status = 'positive'
      }
      if (value === 'disabled') {
        status = 'negative'
      }
      return status
    },
    statusLabel (value) {
      let status = value
      if (value === 'enabled') {
        status = 'success'
      }
      if (value === 'disabled') {
        status = 'error'
      }
      return status
    }
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .panel-service-details {
    height: 600px;
    &-dense{
      height: 400px;
    }
    .q-card__section {
      padding: 24px;
      + .q-card__section {
        padding-top: 0;
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
        &-enabled {
          color: $positive;
        }
        &-disabled {
          color: $negative;
        }
        &-warning {
          color: $warning;
        }
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

    .provider-logo {
      width: 32px;
      height: 32px;
      img {
        width: 100%;
        height: 100%;
      }
    }
  }

</style>
