<template>
  <q-card flat bordered v-bind:class="['panel-router-details']">
    <q-scroll-area :thumb-style="appThumbStyle" style="height:100%;">
      <div v-for="(middleware, index) in data" :key="index">
        <q-card-section class="app-title">
          <div class="app-title-label">{{middleware.type | middlewareTypeLabel}}</div>
        </q-card-section>
        <q-card-section>
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">STATUS</div>
              <div class="block-right-text">
                <avatar-state :state="middleware.status | status "/>
                <div v-bind:class="['block-right-text-label', `block-right-text-label-${middleware.status}`]">{{middleware.status | statusLabel}}</div>
              </div>
            </div>
            <div class="col">
              <div class="text-subtitle2">PROVIDER</div>
              <div class="block-right-text">
                <q-avatar class="provider-logo">
                  <q-icon :name="`img:statics/providers/${middleware.provider}.svg`" />
                </q-avatar>
                <div class="block-right-text-label">{{middleware.provider}}</div>
              </div>
            </div>
          </div>
        </q-card-section>
        <q-card-section v-if="middleware.name">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">NAME</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-name">
                {{ middleware.name }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- TODO - EXTRA FIELDS TO MIDDLEWARES TYPES -->
        <q-card-section v-if="middleware.service">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">SERVICE</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-service">
                {{ middleware.service }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <q-separator v-if="(index+1) < data.length" inset />
      </div>
    </q-scroll-area>
  </q-card>
</template>

<script>
import AvatarState from './AvatarState'

export default {
  name: 'PanelRouterDetails',
  props: ['data'],
  components: {
    AvatarState
  },
  computed: {

  },
  methods: {

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

  .panel-router-details {
    height: 600px;
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
