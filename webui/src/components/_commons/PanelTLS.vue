<template>
  <q-card flat bordered v-bind:class="['panel-tls']">
    <q-scroll-area :thumb-style="appThumbStyle" style="height:100%;">
      <q-card-section v-if="data.options">
        <div class="row items-start no-wrap">
          <div class="col">
            <div class="text-subtitle2">OPTIONS</div>
            <q-chip
              outline
              dense
              class="app-chip app-chip-options">
              {{ data.options }}
            </q-chip>
          </div>
        </div>
      </q-card-section>
      <q-card-section v-if="protocol == 'tcp'">
        <div class="row items-start no-wrap">
          <div class="col">
            <div class="text-subtitle2">PASSTHROUGH</div>
            <boolean-state :value="data.passthrough"></boolean-state>
          </div>
        </div>
      </q-card-section>
      <q-card-section v-if="data.certResolver">
        <div class="row items-start no-wrap">
          <div class="col">
            <div class="text-subtitle2">CERTIFICATE RESOLVER</div>
            <q-chip
              outline
              dense
              class="app-chip app-chip-service">
              {{ data.certResolver }}
            </q-chip>
          </div>
        </div>
      </q-card-section>
      <q-card-section v-if="data.domains">
        <div class="row items-start no-wrap">
          <div class="col">
            <div class="text-subtitle2">DOMAINS</div>
            <div v-for="(domain, key) in data.domains" :key="key" class="flex">
              <q-chip
                outline
                dense
                class="app-chip app-chip-rule">
                {{ domain.main }}
              </q-chip>
              <q-chip
                v-for="(domain, key) in domain.sans" :key="key"
                outline
                dense
                class="app-chip app-chip-entry-points">
                {{ domain }}
              </q-chip>
            </div>
          </div>
        </div>
      </q-card-section>
    </q-scroll-area>
  </q-card>
</template>

<script>
import BooleanState from './BooleanState'

export default {
  name: 'PanelTLS',
  components: {
    BooleanState
  },
  props: ['data', 'protocol']
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .panel-tls {
    height: 600px;
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

    .app-chip {
      &-entry-points {
        display: flex;
        height: 100%;
        flex-wrap: wrap;
        border-width: 0;
        margin-bottom: 8px;
        /deep/ .q-chip__content{
          white-space: normal;
        }
      }
    }
  }

</style>
