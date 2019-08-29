<template>
  <page-default>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-xl q-pb-xl">
        <div v-if="!loading" class="row items-start">

          <div v-if="entryPoints.length" class="col-12 col-md-3 q-mb-lg path-block">
            <div class="row no-wrap items-center q-mb-lg app-title">
              <q-icon name="eva-log-in-outline"></q-icon>
              <div class="app-title-label">Entrypoints</div>
            </div>
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12 col-md-8">
                <div class="row items-start q-col-gutter-md">
                  <div v-for="(entryPoint, index) in entryPoints" :key="index" class="col-12">
                    <panel-entry type="detail" :name="entryPoint.name" :address="entryPoint.address"/>
                  </div>
                </div>
              </div>
              <div class="col-12 col-md-4 xs-hide sm-hide">
                <q-icon name="eva-arrow-forward-outline" class="arrow"></q-icon>
              </div>
            </div>
          </div>

          <div v-if="routerByName.item.name" class="col-12 col-md-3 q-mb-lg path-block">
            <div class="row no-wrap items-center q-mb-lg app-title">
              <q-icon name="eva-globe-outline"></q-icon>
              <div class="app-title-label">HTTP Router</div>
            </div>
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12 col-md-8">
                <div class="row items-start q-col-gutter-md">
                  <div class="col-12">
                    <panel-entry focus="true" type="detail" name="router" :address="routerByName.item.name"/>
                  </div>
                </div>
              </div>
              <div class="col-12 col-md-4 xs-hide sm-hide">
                <q-icon name="eva-arrow-forward-outline" class="arrow"></q-icon>
              </div>
            </div>
          </div>

          <div v-if="routerByName.item.service" class="col-12 col-md-3 q-mb-lg path-block">
            <div class="row no-wrap items-center q-mb-lg app-title">
              <q-icon name="eva-flash"></q-icon>
              <div class="app-title-label">Service</div>
            </div>
            <div class="row items-start q-col-gutter-lg">
              <div class="col-12 col-md-8">
                <div class="row items-start q-col-gutter-md">
                  <div class="col-12">
                    <panel-entry type="detail" name="Service" :address="routerByName.item.service"/>
                  </div>
                </div>
              </div>
            </div>
          </div>

        </div>
        <div v-else class="row items-start">
          <div class="col-12">
            <p v-for="n in 4" :key="n" class="flex">
              <SkeletonBox :min-width="15" :max-width="15" style="margin-right: 2%"/> <SkeletonBox :min-width="50" :max-width="83"/>
            </p>
          </div>
        </div>
      </div>
    </section>

  </page-default>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'
import PageDefault from '../../components/_commons/PageDefault'
import SkeletonBox from '../../components/_commons/SkeletonBox'
import PanelEntry from '../../components/dashboard/PanelEntry'

export default {
  name: 'PageHTTPRoutersDetail',
  props: ['name'],
  components: {
    PageDefault,
    SkeletonBox,
    PanelEntry
  },
  data () {
    return {
      loading: true,
      entryPoints: [],
      timeOutGetAll: null
    }
  },
  computed: {
    ...mapGetters('http', { routerByName: 'routerByName' })
  },
  methods: {
    ...mapActions('http', { getRouterByName: 'getRouterByName' }),
    ...mapActions('entrypoints', { getEntrypointsByName: 'getByName' }),
    refreshAll () {
      if (this.routerByName.loading) {
        return
      }
      this.onGetAll()
    },
    onGetAll () {
      this.getRouterByName(this.name)
        .then(body => {
          if (!body) {
            this.loading = false
            return
          }
          // Get entryPoints
          if (body.entryPoints) {
            for (const entryPoint in body.entryPoints) {
              if (body.entryPoints.hasOwnProperty(entryPoint)) {
                this.getEntrypointsByName(body.entryPoints[entryPoint])
                  .then(body => {
                    if (body) {
                      this.entryPoints.push(body)
                    }
                  })
                  .catch(error => {
                    console.log('Error -> entrypoints/byName', error)
                  })
              }
            }
          }
          clearTimeout(this.timeOutGetAll)
          this.timeOutGetAll = setTimeout(() => {
            this.loading = false
          }, 300)
        })
        .catch(error => {
          console.log('Error -> http/routers/byName', error)
        })
    }
  },
  created () {
    console.log(this.name)
    this.refreshAll()
  },
  mounted () {

  },
  beforeDestroy () {
    clearInterval(this.timeOutGetAll)
    this.$store.commit('http/getRouterByNameClear')
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .path-block {
    .arrow {
      font-size: 40px;
      margin-top: 20px;
      margin-left: 20px;
      color: #b2b2b2;
    }
  }

</style>
