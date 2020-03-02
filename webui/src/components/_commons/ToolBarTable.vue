<template>
  <q-toolbar class="row no-wrap items-center">
    <q-btn-toggle
      v-model="getStatus"
      class="bar-toggle"
      toggle-color="app-toggle"
      text-color="app-grey"
      size="14px"
      no-caps
      rounded
      unelevated
      :options="[
          {label: 'All Status', value: ''},
          {label: 'Success', value: 'enabled'},
          {label: 'Warnings', value: 'warning'},
          {label: 'Errors', value: 'disabled'}
        ]"
    />
    <q-space />
    <q-input v-model="getFilter" rounded dense outlined type="search" debounce="500" placeholder="Search" :bg-color="$q.dark.isActive ? undefined : 'white'" class="bar-search">
      <template v-slot:append>
        <q-icon name="eva-search-outline" />
      </template>
    </q-input>
  </q-toolbar>
</template>

<script>
import Helps from '../../_helpers/Helps'

export default {
  name: 'ToolBarTable',
  props: ['status', 'filter'],
  components: {

  },
  data () {
    return {
    }
  },
  mounted () {
    this.routeToState(this.$route)
  },
  computed: {
    getStatus: {
      get () {
        return this.status
      },
      set (newValue) {
        this.$emit('update:status', newValue)
        this.stateToRoute(this.$route, { status: newValue })
      }
    },
    getFilter: {
      get () {
        return this.filter
      },
      set (newValue) {
        this.$emit('update:filter', newValue)
        this.stateToRoute(this.$route, { filter: newValue })
      }
    }
  },
  watch: {
    $route (to, from) {
      this.routeToState(to)
    }
  },
  methods: {
    routeToState (route) {
      for (const query in route.query) {
        this.$emit(`update:${query}`, route.query[query])
      }
    },
    stateToRoute (route, values) {
      this.$router.push({
        path: route.path,
        query: Helps.removeEmptyObjects({
          ...route.query,
          ...values
        })
      })
    }
  },
  created () {

  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .q-toolbar {
    padding: 0;
    /deep/ .bar-toggle {
      .q-btn {
        font-weight: 600;
        margin-right: 12px;
        &.q-btn--rounded {
          border-radius: 12px;
        }
        &.bg-app-toggle {
          color: $accent !important;
        }
      }
    }
    /deep/ .bar-search {
      .q-field__inner {
        .q-field__control {
          border-radius: 12px;
        }
      }
    }
  }

</style>
