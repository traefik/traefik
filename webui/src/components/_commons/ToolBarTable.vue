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
    <q-input v-model="getFilter" rounded dense outlined type="search" debounce="500" placeholder="Search" bg-color="white" class="bar-search">
      <template v-slot:append>
        <q-icon name="eva-search-outline" />
      </template>
    </q-input>
  </q-toolbar>
</template>

<script>
export default {
  name: 'ToolBarTable',
  props: ['status', 'filter'],
  components: {

  },
  data () {
    return {
    }
  },
  computed: {
    getStatus: {
      get () {
        return this.status
      },
      set (newValue) {
        this.$emit('update:status', newValue)
      }
    },
    getFilter: {
      get () {
        return this.filter
      },
      set (newValue) {
        this.$emit('update:filter', newValue)
      }
    }
  },
  methods: {

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
