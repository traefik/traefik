<template>
  <div class="table-wrapper">
    <q-infinite-scroll @load="handleLoadMore" :offset="250" ref="scroller">
      <q-markup-table>
        <thead>
          <tr class="table-header">
            <th
              v-for="column in getColumns()"
              v-bind:class="`text-${column.align}`"
              v-bind:key="column.name">
              {{ column.label }}
            </th>
          </tr>
        </thead>
        <tfoot v-if="!data.length">
          <tr>
            <td colspan="100%">
              <q-icon name="warning" style="font-size: 1.5rem"/> No data available
            </td>
          </tr>
        </tfoot>
        <tbody>
          <tr v-for="row in data" :key="row.name" @click.native="$router.push({ path: `/${getPath}/${row.name}`})">
            <td v-if="hasColumn('status')" v-bind:class="`text-${getColumn('status').align}`">
              <avatar-state :state="row.status | status "/>
            </td>
            <td v-if="hasColumn('tls')" v-bind:class="`text-${getColumn('tls').align}`">
              <t-l-s-state :is-t-l-s="row.tls"/>
            </td>
            <td v-if="hasColumn('rule')" v-bind:class="`text-${getColumn('rule').align}`">
              <q-chip
                :v-if="row.rule"
                dense
                class="app-chip app-chip-rule">
                {{ row.rule }}
              </q-chip>
            </td>
            <td v-if="hasColumn('entryPoints')" v-bind:class="`text-${getColumn('entryPoints').align}`">
              <div v-if="row.using">
                <q-chip
                  v-for="(entryPoints, index) in row.using" :key="index"
                  dense
                  class="app-chip app-chip-entry-points">
                  {{ entryPoints }}
                </q-chip>
              </div>
            </td>
            <td v-if="hasColumn('name')" v-bind:class="`text-${getColumn('name').align}`">
              <q-chip
                v-if="row.name"
                dense
                class="app-chip app-chip-name">
                {{ row.name }}
              </q-chip>
            </td>
            <td v-if="hasColumn('type')" v-bind:class="`text-${getColumn('type').align}`">
              <q-chip
                v-if="row.type"
                dense
                class="app-chip app-chip-entry-points">
                {{ row.type }}
              </q-chip>
            </td>
            <td v-if="hasColumn('servers')" v-bind:class="`text-${getColumn('servers').align}`">
              <span class="servers-label">{{ row | servers }}</span>
            </td>
            <td v-if="hasColumn('service')" v-bind:class="`text-${getColumn('service').align}`">
              <q-chip
                v-if="row.service"
                dense
                class="app-chip app-chip-service">
                {{ row.service }}
              </q-chip>
            </td>
            <td v-if="hasColumn('provider')" v-bind:class="`text-${getColumn('provider').align}`">
              <q-avatar class="provider-logo">
                <q-icon :name="`img:statics/providers/${row.provider}.svg`" />
              </q-avatar>
            </td>
          </tr>
        </tbody>
      </q-markup-table>
      <template v-slot:loading v-if="loading">
        <div class="row justify-center q-my-md">
          <q-spinner-dots color="app-grey" size="40px" />
        </div>
      </template>
    </q-infinite-scroll>
    <q-page-scroller position="bottom" :scroll-offset="150" class="back-to-top">
      <q-btn color="primary" v-back-to-top small>
        Back to top
      </q-btn>
    </q-page-scroller>
  </div>
</template>

<script>
import AvatarState from './AvatarState'
import TLSState from './TLSState'
import { QMarkupTable, QInfiniteScroll, QSpinnerDots, QPageScroller } from 'quasar'

export default {
  name: 'MainTable',
  props: ['data', 'request', 'loading', 'pagination', 'type', 'onLoadMore', 'endReached'],
  components: {
    TLSState,
    AvatarState,
    QMarkupTable,
    QInfiniteScroll,
    QSpinnerDots,
    QPageScroller
  },
  data () {
    return {
      visibleColumnsHttpRouters: ['status', 'rule', 'entryPoints', 'name', 'service', 'tls', 'provider'],
      visibleColumnsHttpServices: ['status', 'name', 'type', 'servers', 'provider'],
      visibleColumnsHttpMiddlewares: ['status', 'name', 'type', 'provider'],
      visibleColumns: ['status', 'name', 'provider'],
      columns: [
        {
          name: 'status',
          required: true,
          label: 'Status',
          align: 'left',
          field: row => row.status
        },
        {
          name: 'tls',
          align: 'left',
          label: 'TLS',
          field: row => row
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
          name: 'type',
          align: 'left',
          label: 'Type',
          field: row => row.type
        },
        {
          name: 'servers',
          align: 'right',
          label: 'Servers',
          field: row => row.servers
        },
        {
          name: 'service',
          align: 'left',
          label: 'Service',
          field: row => row.service
        },
        {
          name: 'provider',
          align: 'center',
          label: 'Provider',
          field: row => row.provider
        }
      ]
    }
  },
  filters: {
    status (value) {
      if (value === 'enabled') {
        return 'positive'
      }
      if (value === 'disabled') {
        return 'negative'
      }
      return value
    },
    servers (value) {
      if (value.loadBalancer && value.loadBalancer.servers) {
        return value.loadBalancer.servers.length
      }
      return 0
    }
  },
  created () {
    if (this.type === 'http-routers' || this.type === 'tcp-routers') {
      this.visibleColumns = this.visibleColumnsHttpRouters
    }
    if (this.type === 'http-services' || this.type === 'tcp-services') {
      this.visibleColumns = this.visibleColumnsHttpServices
    }
    if (this.type === 'http-middlewares') {
      this.visibleColumns = this.visibleColumnsHttpMiddlewares
    }
  },
  methods: {
    hasColumn (columnName) {
      return this.visibleColumns.includes(columnName)
    },
    getColumns () {
      return this.columns.filter(c => this.visibleColumns.includes(c.name))
    },
    getColumn (columnName) {
      return this.columns.find(c => c.name === columnName) || {}
    },
    handleLoadMore (index, done) {
      this.onLoadMore({ page: index }).then(done).catch(e => {})
    }
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

    .back-to-top {
      margin: 16px 0;
    }
  }

  .servers-label{
    font-size: 14px;
    font-weight: 600;
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
