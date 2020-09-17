<template>
  <page-default>
    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-xl q-pb-lg">
        <div class="row no-wrap items-center q-mb-lg app-title">
          <q-icon name="eva-log-in-outline"></q-icon>
          <div class="app-title-label">Entrypoints</div>
        </div>
        <div v-if="!loadingEntryGetAll" class="row items-center q-col-gutter-lg">
          <div
            v-for="(entryItems, index) in entryAll.items" :key="index"
            class="col-12 col-sm-6 col-md-2">
            <panel-entry :name="entryItems.name" :address="entryItems.address"/>
          </div>
        </div>
        <div v-else class="row items-center q-col-gutter-lg">
          <div class="col-12 col-sm-6 col-md-2">
            <p v-for="n in 3" :key="n" class="flex">
              <SkeletonBox :min-width="15" :max-width="15" style="margin-right: 2%"/> <SkeletonBox :min-width="50" :max-width="83"/>
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-lg q-pb-lg">
        <div class="row no-wrap items-center q-mb-lg app-title">
          <q-icon name="eva-globe-outline"></q-icon>
          <div class="app-title-label">HTTP</div>
        </div>
        <div v-if="!loadingOverview" class="row items-center q-col-gutter-lg">
          <div
            v-for="(overviewHTTP, index) in allHTTP" :key="index"
            class="col-12 col-sm-6 col-md-4">
            <panel-chart :name="index" :data="overviewHTTP" type="http"/>
          </div>
        </div>
        <div v-else class="row items-center q-col-gutter-lg">
          <div class="col-12 col-sm-6 col-md-4">
            <p v-for="n in 6" :key="n" class="flex">
              <SkeletonBox :min-width="15" :max-width="15" style="margin-right: 2%"/> <SkeletonBox :min-width="50" :max-width="83"/>
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-lg q-pb-lg">
        <div class="row no-wrap items-center q-mb-lg app-title">
          <q-icon name="eva-globe-3"></q-icon>
          <div class="app-title-label">TCP</div>
        </div>
        <div v-if="!loadingOverview" class="row items-center q-col-gutter-lg">
          <div
            v-for="(overviewTCP, index) in allTCP" :key="index"
            class="col-12 col-sm-6 col-md-4">
            <panel-chart :name="index" :data="overviewTCP" type="tcp"/>
          </div>
        </div>
        <div v-else class="row items-center q-col-gutter-lg">
          <div class="col-12 col-sm-6 col-md-4">
            <p v-for="n in 6" :key="n" class="flex">
              <SkeletonBox :min-width="15" :max-width="15" style="margin-right: 2%"/> <SkeletonBox :min-width="50" :max-width="83"/>
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-lg q-pb-lg">
        <div class="row no-wrap items-center q-mb-lg app-title">
          <q-icon name="eva-globe-3"></q-icon>
          <div class="app-title-label">UDP</div>
        </div>
        <div v-if="!loadingOverview" class="row items-center q-col-gutter-lg">
          <div
            v-for="(overviewUDP, index) in allUDP" :key="index"
            class="col-12 col-sm-6 col-md-4">
            <panel-chart :name="index" :data="overviewUDP" type="udp"/>
          </div>
        </div>
        <div v-else class="row items-center q-col-gutter-lg">
          <div class="col-12 col-sm-6 col-md-4">
            <p v-for="n in 6" :key="n" class="flex">
              <SkeletonBox :min-width="15" :max-width="15" style="margin-right: 2%"/> <SkeletonBox :min-width="50" :max-width="83"/>
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-lg q-pb-lg">
        <div class="row no-wrap items-center q-mb-lg app-title">
          <q-icon name="eva-toggle-right"></q-icon>
          <div class="app-title-label">Features</div>
        </div>
        <div v-if="!loadingOverview" class="row items-center q-col-gutter-lg">
          <div
            v-for="(overviewFeature, index) in allFeatures" :key="index"
            class="col-12 col-sm-6 col-md-2">
            <panel-feature :feature-key="index" :feature-val="overviewFeature"/>
          </div>
        </div>
        <div v-else class="row items-center q-col-gutter-lg">
          <div class="col-12 col-sm-6 col-md-2">
            <p v-for="n in 3" :key="n" class="flex">
              <SkeletonBox :min-width="15" :max-width="15" style="margin-right: 2%"/> <SkeletonBox :min-width="50" :max-width="83"/>
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-lg q-pb-xl">
        <div class="row no-wrap items-center q-mb-lg app-title">
          <q-icon name="eva-cube"></q-icon>
          <div class="app-title-label">Providers</div>
        </div>
        <div v-if="!loadingOverview" class="row items-center q-col-gutter-lg">
          <div
            v-for="(overviewProvider, index) in allProviders" :key="index"
            class="col-12 col-sm-6 col-md-2">
            <panel-provider :name="overviewProvider"/>
          </div>
        </div>
        <div v-else class="row items-center q-col-gutter-lg">
          <div class="col-12 col-sm-6 col-md-2">
            <p v-for="n in 3" :key="n" class="flex">
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
import PanelChart from '../../components/dashboard/PanelChart'
import PanelFeature from '../../components/dashboard/PanelFeature'
import PanelProvider from '../../components/dashboard/PanelProvider'

export default {
  name: 'PageDashboardIndex',
  components: {
    PageDefault,
    SkeletonBox,
    PanelEntry,
    PanelChart,
    PanelFeature,
    PanelProvider
  },
  data () {
    return {
      loadingEntryGetAll: true,
      loadingOverview: true,
      timeOutEntryGetAll: null,
      timeOutOverviewAll: null,
      intervalRefresh: null,
      intervalRefreshTime: 5000
    }
  },
  computed: {
    ...mapGetters('entrypoints', { entryAll: 'all' }),
    ...mapGetters('core', { overviewAll: 'allOverview' }),
    allHTTP () {
      return this.overviewAll.items.http
    },
    allTCP () {
      return this.overviewAll.items.tcp
    },
    allUDP () {
      return this.overviewAll.items.udp
    },
    allFeatures () {
      return this.overviewAll.items.features
    },
    allProviders () {
      return this.overviewAll.items.providers
    }
  },
  methods: {
    ...mapActions('entrypoints', { entryGetAll: 'getAll' }),
    ...mapActions('core', { getOverview: 'getOverview' }),
    fetchEntries () {
      this.entryGetAll()
        .then(body => {
          console.log('Success -> dashboard/entrypoints', body)
          clearTimeout(this.timeOutEntryGetAll)
          this.timeOutEntryGetAll = setTimeout(() => {
            this.loadingEntryGetAll = false
          }, 300)
        })
        .catch(error => {
          console.log('Error -> dashboard/entrypoints', error)
        })
    },
    fetchOverview () {
      this.getOverview()
        .then(body => {
          console.log('Success -> dashboard/overview', body)
          clearTimeout(this.timeOutOverviewAll)
          this.timeOutOverviewAll = setTimeout(() => {
            this.loadingOverview = false
          }, 300)
        })
        .catch(error => {
          console.log('Error -> dashboard/overview', error)
        })
    },
    fetchAll () {
      this.fetchEntries()
      this.fetchOverview()
    }
  },
  created () {
    this.fetchAll()
    this.intervalRefresh = setInterval(this.fetchOverview, this.intervalRefreshTime)
  },
  beforeDestroy () {
    clearInterval(this.intervalRefresh)
    clearTimeout(this.timeOutEntryGetAll)
    clearTimeout(this.timeOutOverviewAll)
    this.$store.commit('entrypoints/getAllClear')
    this.$store.commit('core/getOverviewClear')
  }
}
</script>

<style scoped lang="scss">

</style>
