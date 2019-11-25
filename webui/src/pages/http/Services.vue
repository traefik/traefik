<template>
  <page-default>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-xl q-pb-xl">
        <div class="row no-wrap items-center q-mb-lg">
          <tool-bar-table :status.sync="status" :filter.sync="filter"/>
        </div>
        <div class="row items-center q-col-gutter-lg">
          <div class="col-12">
            <main-table
              ref="mainTable"
              v-bind="getTableProps({ type: 'http-services' })"
              :data="allServices.items"
              :onLoadMore="handleLoadMore"
              :endReached="allServices.endReached"
              :loading="allServices.loading"
            />
          </div>
        </div>
      </div>
    </section>

  </page-default>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'
import GetTablePropsMixin from '../../_mixins/GetTableProps'
import PageDefault from '../../components/_commons/PageDefault'
import ToolBarTable from '../../components/_commons/ToolBarTable'
import MainTable from '../../components/_commons/MainTable'

export default {
  name: 'PageHTTPServices',
  mixins: [GetTablePropsMixin],
  components: {
    PageDefault,
    ToolBarTable,
    MainTable
  },
  data () {
    return {
      loading: true,
      filter: '',
      status: '',
      pagination: {
        sortBy: '',
        descending: true,
        page: 1,
        rowsPerPage: 10,
        rowsNumber: 0
      }
    }
  },
  computed: {
    ...mapGetters('http', { allServices: 'allServices' })
  },
  methods: {
    ...mapActions('http', { getAllServices: 'getAllServices' }),
    initData () {
      const scrollerRef = this.$refs.mainTable.$refs.scroller
      if (scrollerRef) {
        scrollerRef.stop()
        scrollerRef.reset()
      }

      this.handleLoadMore({ page: 1 }).then(() => {
        if (scrollerRef) {
          scrollerRef.resume()
          scrollerRef.poll()
        }
      })
    },
    refreshAll () {
      if (this.allServices.loading) {
        return
      }

      this.handleLoadMore({ page: 1 })
    },
    handleLoadMore ({ page = 1 } = {}) {
      this.pagination.page = page
      return this.onGetAll({
        pagination: this.pagination,
        filter: this.filter
      })
    },
    onGetAll (props) {
      let { page, rowsPerPage, sortBy, descending } = props.pagination

      return this.getAllServices({ query: props.filter, status: this.status, page, limit: rowsPerPage, sortBy, descending })
        .then(body => {
          this.loading = false
          console.log('Success -> http/services', body)
          // update rowsNumber with appropriate value
          this.pagination.rowsNumber = body.total
          // update local pagination object
          this.pagination.page = page
          this.pagination.rowsPerPage = rowsPerPage
          this.pagination.sortBy = sortBy
          this.pagination.descending = descending
        })
    }
  },
  watch: {
    'status' () {
      this.initData()
    },
    'filter' () {
      this.initData()
    }
  },
  created () {

  },
  mounted () {
    this.refreshAll()
  },
  beforeDestroy () {
    this.$store.commit('http/getAllServicesClear')
  }
}
</script>

<style scoped lang="scss">

</style>
