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
        <q-tr :props="props" class="cursor-pointer" @click.native="$router.push({ path: `/${getPath}/${props.row.name}/${getType(props.row)}`})">
          <q-td key="status" :props="props" auto-width>
            <avatar-state :state="props.row.status | status "/>
          </q-td>
          <q-td key="tls" :props="props" auto-width>
            <t-l-s-state :is-t-l-s="props.row.tls"/>
          </q-td>
          <q-td key="rule" :props="props">
            <q-chip
              v-if="props.row.rule"
              outline
              dense
              class="app-chip app-chip-rule">
              {{ props.row.rule }}
            </q-chip>
          </q-td>
          <q-td key="entryPoints" :props="props">
            <div v-if="props.row.using">
              <q-chip
                v-for="(entryPoints, index) in props.row.using" :key="index"
                outline
                dense
                class="app-chip app-chip-entry-points">
                {{ entryPoints }}
              </q-chip>
            </div>
          </q-td>
          <q-td key="name" :props="props">
            <q-chip
              v-if="props.row.name"
              outline
              dense
              class="app-chip app-chip-name">
              {{ props.row.name }}
            </q-chip>
          </q-td>
          <q-td key="type" :props="props">
            <q-chip
              v-if="props.row.type"
              outline
              dense
              class="app-chip app-chip-entry-points">
              {{ props.row.type | capFirstLetter}}
            </q-chip>
          </q-td>
          <q-td key="servers" :props="props">
            <span class="servers-label">{{ props.row | servers }}</span>
          </q-td>
          <q-td key="service" :props="props">
            <q-chip
              v-if="props.row.service"
              outline
              dense
              class="app-chip app-chip-service">
              {{ props.row.service }}
            </q-chip>
          </q-td>
          <q-td key="provider" :props="props" auto-width>
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
import TLSState from './TLSState'

export default {
  name: 'MainTable',
  props: ['data', 'request', 'loading', 'pagination', 'filter', 'type'],
  components: {
    TLSState,
    AvatarState
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
  computed: {
    syncPagination: {
      get () {
        return this.pagination
      },
      set (newValue) {
        this.$emit('update:pagination', newValue)
      }
    },
    getPath () {
      return this.type.replace('-', '/', 'gi')
    }
  },
  methods: {
    getType (item) {
      return item.type || 'default'
    }
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
    },
    servers (value) {
      let servers = 0
      if (value.loadBalancer && value.loadBalancer.servers) {
        servers = value.loadBalancer.servers.length
      }
      return servers
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
