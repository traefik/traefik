<template>
  <page-default>
    <section
      v-if="!loading"
      class="app-section"
    >
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-xl q-pb-sm">
        <div
          v-if="middlewareByName.item"
          class="row no-wrap items-center app-title"
        >
          <div
            class="app-title-label"
            style="font-size: 26px"
          >
            {{ middlewareByName.item.name }}
          </div>
        </div>
      </div>
    </section>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-sm q-pb-lg">
        <div
          v-if="!loading"
          class="row items-start q-col-gutter-md"
        >
          <div
            v-if="middlewareByName.item"
            class="col-12 col-md-4 q-mb-lg path-block"
          >
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12">
                <div class="row items-start q-col-gutter-md">
                  <div class="col-12">
                    <panel-middlewares
                      dense
                      :data="[middlewareByName.item]"
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
import PanelMiddlewares from '../../components/_commons/PanelMiddlewares.vue'
import MainTable from '../../components/_commons/MainTable.vue'

export default defineComponent({
  name: 'PageMiddlewareDetail',
  components: {
    PageDefault,
    SkeletonBox,
    PanelMiddlewares,
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
    ...mapGetters('http', { http_middlewareByName: 'middlewareByName' }),
    ...mapGetters('tcp', { tcp_middlewareByName: 'middlewareByName' }),
    protocol () {
      return this.$route.meta.protocol
    },
    middlewareByName () {
      return this[`${this.protocol}_middlewareByName`]
    },
    getMiddlewareByName () {
      return this[`${this.protocol}_getMiddlewareByName`]
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
    this.$store.commit('http/getMiddlewareByNameClear')
    this.$store.commit('tcp/getMiddlewareByNameClear')
  },
  methods: {
    ...mapActions('http', { http_getMiddlewareByName: 'getMiddlewareByName', http_getRouterByName: 'getRouterByName', http_getAllRouters: 'getAllRouters' }),
    ...mapActions('tcp', { tcp_getMiddlewareByName: 'getMiddlewareByName', tcp_getRouterByName: 'getRouterByName', tcp_getAllRouters: 'getAllRouters' }),
    refreshAll () {
      if (this.middlewareByName.loading) {
        return
      }
      this.onGetAll()
    },
    onGetAll () {
      this.getMiddlewareByName(this.name)
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
            middlewareName: this.name,
            serviceName: '',
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
              console.log('Error -> routers/byName', error)
            })
          clearTimeout(this.timeOutGetAll)
          this.timeOutGetAll = setTimeout(() => {
            this.loading = false
          }, 300)
        })
        .catch(error => {
          console.log('Error -> middleware/byName', error)
        })
    }
  }
})
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

</style>
