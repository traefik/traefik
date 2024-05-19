<template>
  <page-default>
    <section
      v-if="!loading"
      class="app-section"
    >
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-xl q-pb-lg">
        <div
          v-if="serviceByName.item"
          class="row no-wrap items-center app-title"
        >
          <div
            class="app-title-label"
            style="font-size: 26px"
          >
            {{ serviceByName.item.name }}
          </div>
        </div>
      </div>
    </section>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-lg q-pb-lg">
        <div
          v-if="!loading"
          class="row items-start q-col-gutter-md"
        >
          <div
            v-if="serviceByName.item"
            class="col-12 col-md-4 q-mb-lg path-block"
          >
            <div class="row no-wrap items-center q-mb-lg app-title">
              <q-icon name="eva-info" />
              <div class="app-title-label">
                Service Details
              </div>
            </div>
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12">
                <div class="row items-start q-col-gutter-md">
                  <div class="col-12">
                    <panel-service-details
                      dense
                      :data="serviceByName.item"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="serviceByName.item.loadBalancer && serviceByName.item.loadBalancer.healthCheck"
            class="col-12 col-md-4 q-mb-lg path-block"
          >
            <div class="row no-wrap items-center q-mb-lg app-title">
              <q-icon name="eva-shield" />
              <div class="app-title-label">
                Health Check
              </div>
            </div>
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12">
                <div class="row items-start q-col-gutter-md">
                  <div class="col-12">
                    <panel-health-check
                      dense
                      :data="serviceByName.item.loadBalancer.healthCheck"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="serviceByName.item.loadBalancer"
            class="col-12 col-md-4 q-mb-lg path-block"
          >
            <div class="row no-wrap items-center q-mb-lg app-title">
              <q-icon name="eva-globe-outline" />
              <div class="app-title-label">
                Servers
              </div>
            </div>
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12">
                <div class="row items-start q-col-gutter-md">
                  <div class="col-12">
                    <panel-servers
                      dense
                      :data="serviceByName.item"
                      :has-status="serviceByName.item.serverStatus"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="serviceByName.item.weighted && serviceByName.item.weighted.services"
            class="col-12 col-md-4 q-mb-lg path-block"
          >
            <div class="row no-wrap items-center q-mb-lg app-title">
              <q-icon name="eva-globe-outline" />
              <div class="app-title-label">
                Services
              </div>
            </div>
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12">
                <div class="row items-start q-col-gutter-md">
                  <div class="col-12">
                    <panel-weighted-services
                      dense
                      :data="serviceByName.item"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="serviceByName.item.mirroring && serviceByName.item.mirroring.mirrors"
            class="col-12 col-md-4 q-mb-lg path-block"
          >
            <div class="row no-wrap items-center q-mb-lg app-title">
              <q-icon name="eva-globe-outline" />
              <div class="app-title-label">
                Mirror Services
              </div>
            </div>
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12">
                <div class="row items-start q-col-gutter-md">
                  <div class="col-12">
                    <panel-mirroring-services
                      dense
                      :data="serviceByName.item"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div
          v-else
          class="row items-start q-mt-xl"
        >
          <div class="col-12">
            <p
              v-for="n in 4"
              :key="n"
              class="flex"
            >
              <SkeletonBox
                :min-width="15"
                :max-width="15"
                style="margin-right: 2%"
              /> <SkeletonBox
                :min-width="50"
                :max-width="83"
              />
            </p>
          </div>
        </div>
      </div>
    </section>

    <section
      v-if="!loading && allRouters.length"
      class="app-section"
    >
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-lg q-pb-xl">
        <div class="row no-wrap items-center q-mb-lg app-title">
          <div class="app-title-label">
            Used by Routers
          </div>
        </div>
        <div class="row items-center q-col-gutter-lg">
          <div class="col-12">
            <main-table
              v-bind="getTableProps({ type: `${protocol}-routers` })"
              v-model:current-sort="sortBy"
              v-model:current-sort-dir="sortDir"
              :data="allRouters"
              :on-load-more="onGetAll"
              :request="()=>{}"
              :loading="routersLoading"
              :filter="routersFilter"
            />
          </div>
        </div>
      </div>
    </section>
  </page-default>
</template>

<script>
import { defineComponent } from 'vue'
import { mapActions, mapGetters } from 'vuex'
import GetTablePropsMixin from '../../_mixins/GetTableProps'
import PageDefault from '../../components/_commons/PageDefault.vue'
import SkeletonBox from '../../components/_commons/SkeletonBox.vue'
import PanelServiceDetails from '../../components/_commons/PanelServiceDetails.vue'
import PanelHealthCheck from '../../components/_commons/PanelHealthCheck.vue'
import PanelServers from '../../components/_commons/PanelServers.vue'
import MainTable from '../../components/_commons/MainTable.vue'
import PanelWeightedServices from '../../components/_commons/PanelWeightedServices.vue'
import PanelMirroringServices from '../../components/_commons/PanelMirroringServices.vue'

export default defineComponent({
  name: 'PageServiceDetail',
  components: {
    PanelMirroringServices,
    PanelWeightedServices,
    PageDefault,
    SkeletonBox,
    PanelServiceDetails,
    PanelHealthCheck,
    PanelServers,
    MainTable
  },
  mixins: [GetTablePropsMixin],
  props: {
    name: {
      default: '',
      type: String
    },
    type: {
      default: '',
      type: String
    }
  },
  data () {
    return {
      loading: true,
      timeOutGetAll: null,
      allRouters: [],
      routersLoading: true,
      routersFilter: '',
      routersStatus: '',
      routersPagination: {
        sortBy: '',
        descending: true,
        page: 1,
        rowsPerPage: 1000,
        rowsNumber: 0
      },
      filter: '',
      status: '',
      sortBy: 'name',
      sortDir: 'asc'
    }
  },
  computed: {
    ...mapGetters('http', { http_serviceByName: 'serviceByName' }),
    ...mapGetters('tcp', { tcp_serviceByName: 'serviceByName' }),
    ...mapGetters('udp', { udp_serviceByName: 'serviceByName' }),
    protocol () {
      return this.$route.meta.protocol
    },
    serviceByName () {
      return this[`${this.protocol}_serviceByName`]
    },
    getServiceByName () {
      return this[`${this.protocol}_getServiceByName`]
    },
    getRouterByName () {
      return this[`${this.protocol}_getRouterByName`]
    },
    getAllRouters () {
      return this[`${this.protocol}_getAllRouters`]
    }
  },
  watch: {
    'sortBy' () {
      this.refreshAll()
    },
    'sortDir' () {
      this.refreshAll()
    }
  },
  created () {
    this.refreshAll()
  },
  mounted () {},
  beforeUnmount () {
    clearInterval(this.timeOutGetAll)
    this.$store.commit('http/getServiceByNameClear')
    this.$store.commit('tcp/getServiceByNameClear')
    this.$store.commit('udp/getServiceByNameClear')
  },
  methods: {
    ...mapActions('http', { http_getServiceByName: 'getServiceByName', http_getRouterByName: 'getRouterByName', http_getAllRouters: 'getAllRouters' }),
    ...mapActions('tcp', { tcp_getServiceByName: 'getServiceByName', tcp_getRouterByName: 'getRouterByName', tcp_getAllRouters: 'getAllRouters' }),
    ...mapActions('udp', { udp_getServiceByName: 'getServiceByName', udp_getRouterByName: 'getRouterByName', udp_getAllRouters: 'getAllRouters' }),
    refreshAll () {
      if (this.serviceByName.loading) {
        return
      }
      this.onGetAll()
    },
    onGetAll () {
      this.getServiceByName(this.name)
        .then(body => {
          if (!body) {
            this.loading = false
            return
          }
          // Get routers
          this.getAllRouters({
            query: this.filter,
            status: this.status,
            page: 1,
            limit: 1000,
            middlewareName: '',
            serviceName: this.name,
            sortBy: this.sortBy,
            direction: this.sortDir
          })
            .then(body => {
              this.allRouters = []
              if (body) {
                this.routersLoading = false
                this.allRouters.push(...body.data)
              }
            })
            .catch(error => {
              console.log('Error -> getAllRouters', error)
            })
          clearTimeout(this.timeOutGetAll)
          this.timeOutGetAll = setTimeout(() => {
            this.loading = false
          }, 300)
        })
        .catch(error => {
          console.log('Error -> service/byName', error)
        })
    }
  }
})
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

</style>
