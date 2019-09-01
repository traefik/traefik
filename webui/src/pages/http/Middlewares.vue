<template>
  <page-default>

    <section class="app-section">
      <div class="app-section-wrap app-boxed app-boxed-xl q-pl-md q-pr-md q-pt-xl q-pb-xl">
        <div class="row no-wrap items-center q-mb-lg">
          <tool-bar-table :status.sync="status" :filter.sync="filter"/>
        </div>
        <div class="row items-center q-col-gutter-lg">
          <div class="col-12">
            <main-table :data="allMiddlewares.items" :request="onGetAll" :loading="loading" :pagination.sync="pagination" :filter="filter" type="http-middlewares"/>
          </div>
        </div>
      </div>
    </section>

  </page-default>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'
import PageDefault from '../../components/_commons/PageDefault'
import ToolBarTable from '../../components/_commons/ToolBarTable'
import MainTable from '../../components/_commons/MainTable'

export default {
  name: 'PageHTTPMiddlewares',
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
    ...mapGetters('http', { allMiddlewares: 'allMiddlewares' })
  },
  methods: {
    ...mapActions('http', { getAllMiddlewares: 'getAllMiddlewares' }),
    refreshAll () {
      if (this.allMiddlewares.loading) {
        return
      }
      this.pagination.page = 1
      this.onGetAll({
        pagination: this.pagination,
        filter: this.filter
      })
    },
    onGetAll (props) {
      let { page, rowsPerPage, sortBy, descending } = props.pagination

      this.getAllMiddlewares({ query: props.filter, status: this.status, page, limit: rowsPerPage, sortBy, descending })
        .then(body => {
          if (!body) {
            this.loading = false
            return
          }
          this.loading = false
          console.log('Success -> http/middlewares', body)
          // update rowsNumber with appropriate value
          this.pagination.rowsNumber = body.total
          // update local pagination object
          this.pagination.page = page
          this.pagination.rowsPerPage = rowsPerPage
          this.pagination.sortBy = sortBy
          this.pagination.descending = descending
        })
        .catch(error => {
          console.log('Error -> http/middlewares', error)
        })
    }
  },
  watch: {
    'status' () {
      this.refreshAll()
    },
    'filter' () {
      this.refreshAll()
    }
  },
  created () {

  },
  mounted () {
    this.refreshAll()
  },
  beforeDestroy () {
    this.$store.commit('http/getAllMiddlewaresClear')
  }
}
</script>

<style scoped lang="scss">

</style>
