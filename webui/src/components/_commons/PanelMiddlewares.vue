<template>
  <q-card flat bordered v-bind:class="['panel-middleware-details', {'panel-middleware-details-dense':isDense}]">
    <q-scroll-area v-if="data && data.length" :thumb-style="appThumbStyle" style="height:100%;">
      <div v-for="(middleware, index) in data" :key="index">
        <q-card-section v-if="!isDense" class="app-title">
          <div class="app-title-label text-capitalize">{{middleware.type}}</div>
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
    <q-card-section v-else style="height: 100%">
      <div class="row items-center" style="height: 100%">
        <div class="col-12">
        <div class="block-empty"></div>
          <div class="q-pb-lg block-empty-logo">
            <img alt="empty" src="~assets/middlewares-empty.svg">
          </div>
          <div class="block-empty-label">There are no<br>Middlewares configured</div>
        </div>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
import AvatarState from './AvatarState'

export default {
  name: 'PanelMiddlewareDetails',
  props: ['data', 'dense'],
  components: {
    AvatarState
  },
  computed: {
    isDense () {
      return this.dense !== undefined
    }
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

  .panel-middleware-details {
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

    .block-empty {
      &-logo {
        text-align: center;
      }
      &-label {
        font-size: 20px;
        font-weight: 700;
        color: #b8b8b8;
        text-align: center;
        line-height: 1.2;
      }
    }
  }

</style>
