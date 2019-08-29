<template>
  <div class="table-wrapper">
    <q-table
      :data="data"
      :columns="columns"
      row-key="name"
      :pagination.sync="syncPagination"
      :rows-per-page-options="[10, 20, 40, 80, 100, 0]"
      :loading="loading"
      :filter="filter"
      @request="request"
      binary-state-sort
      :visible-columns="visibleColumns"
      color="primary"
      table-header-class="table-header">

      <template v-slot:body="props">
        <q-tr :props="props">
          <q-td key="status" :props="props">
            <avatar-state :state="props.row.status | status "/>
          </q-td>
          <q-td key="rule" :props="props">
            <q-chip
              v-if="props.row.rule"
              outline
              dense
              class="chip-table chip-table-rule">
              {{ props.row.rule }}
            </q-chip>
          </q-td>
          <q-td key="entryPoints" :props="props">
            <div v-if="props.row.entryPoints">
              <q-chip
                v-for="(entryPoints, index) in props.row.entryPoints" :key="index"
                outline
                dense
                class="chip-table chip-table-entry-points">
                {{ entryPoints }}
              </q-chip>
            </div>
          </q-td>
          <q-td key="name" :props="props">
            <q-chip
              v-if="props.row.name"
              outline
              dense
              class="chip-table chip-table-name">
              {{ props.row.name }}
            </q-chip>
          </q-td>
          <q-td key="provider" :props="props">
            <q-avatar class="provider-logo">
              <q-icon :name="`img:statics/providers/${props.row.provider}.svg`" />
            </q-avatar>
          </q-td>
        </q-tr>
      </template>

    </q-table>
  </div>
</template>

<script>
import AvatarState from './AvatarState'

export default {
  name: 'MainTable',
  props: ['data', 'request', 'loading', 'pagination', 'filter', 'type'],
  components: {
    AvatarState
  },
  data () {
    return {
      visibleColumns: ['status', 'rule', 'entryPoints', 'name', 'provider'],
      columns: [
        {
          name: 'status',
          required: true,
          label: 'Status',
          align: 'left',
          field: row => row.status
        },
        {
          name: 'rule',
          align: 'left',
          label: 'Rule',
          field: row => row.rule
        },
        {
          name: 'entryPoints',
          align: 'left',
          label: 'Entrypoints',
          field: row => row.entryPoints
        },
        {
          name: 'name',
          align: 'left',
          label: 'Name',
          field: row => row.name
        },
        {
          name: 'provider',
          align: 'right',
          label: 'Provider',
          field: row => row.provider
        }
      ]
    }
  },
  computed: {
    syncPagination: {
      get () {
        return this.pagination
      },
      set (newValue) {
        this.$emit('update:pagination', newValue)
      }
    }
  },
  methods: {

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
    }
  },
  created () {

  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .table-wrapper {
    /deep/ .q-table__container{
      border-radius: 8px;
      .q-table {
        .table-header {
          th {
            font-size: 14px;
            font-weight: 700;
          }
        }
        tbody {
          tr:hover {
            background: rgba( $accent, 0.04 );
          }
        }
      }
      .q-table__bottom {
        > .q-table__control {
          &:nth-last-child(2) {
            display: none;
          }
          &:nth-last-child(1) {
            .q-table__bottom-item {
              display: none;
            }
          }
        }
      }
    }
  }

  .chip-table {
    border-radius: 8px;
    font-weight: 600;
    font-size: 14px;
    &-rule {
      color: $accent;
      border: solid 1px $accent;
      background-color: rgba($accent, 0.1);
    }
    &-entry-points {
      color: $app-text-green;
      border: solid 1px $app-text-green;
      background-color: rgba($app-text-green, 0.1);
    }
    &-name {
      color: $app-text-purple;
      border: solid 1px $app-text-purple;
      background-color: rgba($app-text-purple, 0.1);
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
</style>
