<template>
  <q-card flat bordered v-bind:class="['panel-middleware-details', {'panel-middleware-details-dense':isDense}]">
    <q-scroll-area v-if="data && data.length" :thumb-style="appThumbStyle" style="height:100%;">
      <div v-for="(middleware, index) in data" :key="index">
        <q-card-section v-if="!isDense" class="app-title">
          <div class="app-title-label text-capitalize">{{ middleware.name }}</div>
        </q-card-section>
        <!-- COMMON FIELDS -->
        <q-card-section>
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Type</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-purple">
                {{ middleware.type }}
              </q-chip>
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
        <q-card-section>
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">STATUS</div>
              <div class="block-right-text">
                <avatar-state :state="middleware.status | status "/>
                <div v-bind:class="['block-right-text-label', `block-right-text-label-${middleware.status}`]">{{middleware.status | statusLabel}}</div>
              </div>
            </div>
          </div>
        </q-card-section>
        <!-- ERROR -->
        <q-card-section v-if="middleware.error">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">ERRORS</div>
              <q-chip
                v-for="(errorMsg, index) in middleware.error" :key="index"
                outline
                class="app-chip app-chip-error">
                {{ errorMsg }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [addPrefix] - prefix -->
        <q-card-section v-if="middleware.addPrefix">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">PREFIX</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).prefix }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [basicAuth & digestAuth] - users -->
        <q-card-section v-if="exData(middleware).users">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">USERS</div>
              <q-chip
                v-for="(user, key) in exData(middleware).users" :key="key"
                outline
                dense
                class="app-chip app-chip-green">
                {{ user }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [basicAuth & digestAuth] - usersFile -->
        <q-card-section v-if="exData(middleware).usersFile">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Users File</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).usersFile }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [basicAuth & digestAuth] - realm -->
        <q-card-section v-if="exData(middleware).realm">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Realm</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-warning">
                {{ exData(middleware).realm }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [basicAuth & digestAuth] - removeHeader -->
        <q-card-section v-if="middleware.basicAuth || middleware.digestAuth">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Remove Header</div>
              <boolean-state :value="exData(middleware).removeHeader"/>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [basicAuth & digestAuth] - headerField -->
        <q-card-section v-if="exData(middleware).headerField">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Header Field</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-warning">
                {{ exData(middleware).headerField }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [chain] - middlewares -->
        <q-card-section v-if="middleware.chain">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Chain</div>
              <q-chip
                v-for="(mi, key) in exData(middleware).middlewares" :key="key"
                outline
                dense
                class="app-chip app-chip-green">
                {{ mi }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [buffering] - xxxRequestBodyBytes -->
        <q-card-section v-if="exData(middleware).maxRequestBodyBytes || exData(middleware).memRequestBodyBytes">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Max Request Body Bytes</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).maxRequestBodyBytes }}
              </q-chip>
            </div>
            <div class="col">
              <div class="text-subtitle2">Mem Request Body Bytes</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).memRequestBodyBytes }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [buffering] - xxxResponseBodyBytes -->
        <q-card-section v-if="exData(middleware).maxResponseBodyBytes || exData(middleware).memResponseBodyBytes">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Max Response Body Bytes</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).maxResponseBodyBytes }}
              </q-chip>
            </div>
            <div class="col">
              <div class="text-subtitle2">Mem Response Body Bytes</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).memResponseBodyBytes }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [buffering] - retryExpression -->
        <q-card-section v-if="exData(middleware).retryExpression">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Retry Expression</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).retryExpression }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [circuitBreaker] - expression -->
        <q-card-section v-if="middleware.circuitBreaker">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Expression</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).expression }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [compress] - compress -->
        <q-card-section v-if="middleware.compress">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Compress</div>
              <boolean-state :value="!!middleware.compress"/>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [errors] - service -->
        <q-card-section v-if="middleware.errors">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Service</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).service }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [errors] - query -->
        <q-card-section v-if="middleware.errors">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Query</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).query }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [errors] - status -->
        <q-card-section v-if="middleware.errors">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Status</div>
              <q-chip
                v-for="(st, key) in exData(middleware).status" :key="key"
                outline
                dense
                class="app-chip app-chip-green">
                {{ st }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [forwardAuth] - address -->
        <q-card-section v-if="middleware.forwardAuth">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Address</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).address }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [forwardAuth] - trustForwardHeader && tls -->
        <q-card-section v-if="middleware.forwardAuth">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">TLS</div>
              <boolean-state :value="!!exData(middleware).tls"/>
            </div>
            <div class="col">
              <div class="text-subtitle2">Trust Forward Headers</div>
              <boolean-state :value="exData(middleware).trustForwardHeader"/>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [forwardAuth] - authResponseHeaders -->
        <q-card-section v-if="middleware.forwardAuth">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Auth Response Headers</div>
              <q-chip
                v-for="(respHeader, key) in exData(middleware).authResponseHeaders" :key="key"
                outline
                dense
                class="app-chip app-chip-green">
                {{ respHeader }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [inflightreq] - amount -->
        <q-card-section v-if="exData(middleware).amount">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">AMOUNT</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-warning">
                {{ exData(middleware).amount }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [inflightreq] - aipStrategy -->
        <q-card-section v-if="exData(middleware).sourceCriterion && exData(middleware).sourceCriterion.ipStrategy">
          <div class="row items-start">
            <div class="col-12">
              <div class="text-subtitle2">IP STRATEGY</div>
            </div>
            <div v-if="exData(middleware).sourceCriterion.ipStrategy.depth" class="col-12">
              <q-chip
                outline
                dense
                class="app-chip app-chip-accent">Depth :</q-chip>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
               {{ exData(middleware).sourceCriterion.ipStrategy.depth }}
              </q-chip>
            </div>
            <div v-if="exData(middleware).sourceCriterion.ipStrategy.excludedIPs" class="col-12">
              <div class="flex">
                <q-chip
                  outline
                  dense
                  class="app-chip app-chip-accent">
                  Excluded IPs:
                </q-chip>
                <q-chip
                  v-for="(excludedIPs, key) in exData(middleware).sourceCriterion.ipStrategy.excludedIPs" :key="key"
                  outline
                  dense
                  class="app-chip app-chip-green">
                  {{ excludedIPs }}
                </q-chip>
              </div>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [inflightreq] - arequestHeaderName, requestHost -->
        <q-card-section v-if="exData(middleware) && exData(middleware).sourceCriterion">
          <div class="row items-start no-wrap">
            <div v-if="exData(middleware).sourceCriterion.requestHeaderName" class="col">
              <div class="text-subtitle2">REQUEST HEADER NAME</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-warning">
                {{ exData(middleware).sourceCriterion.requestHeaderName }}
              </q-chip>
            </div>
            <div v-if="exData(middleware).sourceCriterion.requestHost" class="col">
              <div class="text-subtitle2">REQUEST HOST</div>
              <boolean-state :value="exData(middleware).sourceCriterion.requestHost"/>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [redirectRegex] - regex -->
        <q-card-section v-if="middleware.redirectRegex">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Regex</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).regex }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [redirectRegex] - replacement -->
        <q-card-section v-if="middleware.redirectRegex">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Replacement</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).replacement }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [redirectRegex] - permanent -->
        <q-card-section v-if="middleware.redirectRegex">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Permanent</div>
              <boolean-state :value="!!exData(middleware).permanent"/>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [redirectScheme] - scheme -->
        <q-card-section v-if="middleware.redirectScheme">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Scheme</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).scheme }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [replacePath] - path -->
        <q-card-section v-if="middleware.replacePath">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Path</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).path }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [replacePathRegex] - regex -->
        <q-card-section v-if="middleware.replacePathRegex">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Regex</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).regex }}
              </q-chip>
            </div>
          </div>
        </q-card-section>
        <!-- EXTRA FIELDS FROM MIDDLEWARES - [replacePathRegex] - replacement -->
        <q-card-section v-if="middleware.replacePathRegex">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Replacement</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).replacement }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [retry] - attempts -->
        <q-card-section v-if="middleware.retry">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Attempts</div>
              <q-chip
                outline
                dense
                class="app-chip app-chip-green">
                {{ exData(middleware).attempts }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [stripPrefix] - prefixes -->
        <q-card-section v-if="middleware.stripPrefix">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Prefixes</div>
              <q-chip
                v-for="(prefix, key) in exData(middleware).prefixes" :key="key"
                outline
                dense
                class="app-chip app-chip-green">
                {{ prefix }}
              </q-chip>
            </div>
          </div>
        </q-card-section>

        <!-- EXTRA FIELDS FROM MIDDLEWARES - [stripPrefixRegex] - regex -->
        <q-card-section v-if="middleware.stripPrefixRegex">
          <div class="row items-start no-wrap">
            <div class="col">
              <div class="text-subtitle2">Regex</div>
              <q-chip
                v-for="(exp, key) in exData(middleware).regex" :key="key"
                outline
                dense
                class="app-chip app-chip-green">
                {{ exp }}
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
import BooleanState from './BooleanState'
import AvatarState from './AvatarState'

export default {
  name: 'PanelMiddlewareDetails',
  props: ['data', 'dense'],
  components: {
    AvatarState,
    BooleanState
  },
  computed: {
    isDense () {
      return this.dense !== undefined
    }
  },
  methods: {
    exData (item) {
      let exData = {}
      for (const prop in item) {
        if (prop.toLowerCase() === item.type && item.hasOwnProperty(prop)) {
          exData = item[prop]
        }
      }
      return exData
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

  .panel-middleware-details {
    height: 600px;
    &-dense{
      /*height: 400px;*/
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

    .app-chip {
      &-error {
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
