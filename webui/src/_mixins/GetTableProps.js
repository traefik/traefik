import { getProperty } from 'dot-prop'
import { QChip } from 'quasar'
import Chips from '../components/_commons/Chips.vue'
import ProviderIcon from '../components/_commons/ProviderIcon.vue'
import AvatarState from '../components/_commons/AvatarState.vue'
import TLSState from '../components/_commons/TLSState.vue'

const allColumns = [
  {
    name: 'status',
    required: true,
    label: 'Status',
    align: 'left',
    sortable: true,
    fieldToProps: row => ({
      state: row.status === 'enabled' ? 'positive' : 'negative'
    }),
    component: AvatarState
  },
  {
    name: 'tls',
    align: 'left',
    label: 'TLS',
    sortable: false,
    fieldToProps: row => ({ isTLS: row.tls }),
    component: TLSState
  },
  {
    name: 'rule',
    align: 'left',
    label: 'Rule',
    sortable: true,
    component: QChip,
    fieldToProps: () => ({ class: 'app-chip app-chip-rule', dense: true }),
    content: row => row.rule
  },
  {
    name: 'entryPoints',
    align: 'left',
    label: 'Entrypoints',
    sortable: true,
    component: Chips,
    fieldToProps: row => ({
      classNames: 'app-chip app-chip-entry-points',
      dense: true,
      list: row.entryPoints
    })
  },
  {
    name: 'name',
    align: 'left',
    label: 'Name',
    sortable: true,
    component: QChip,
    fieldToProps: () => ({ class: 'app-chip app-chip-name', dense: true }),
    content: row => row.name
  },
  {
    name: 'type',
    align: 'left',
    label: 'Type',
    sortable: true,
    component: QChip,
    fieldToProps: () => ({
      class: 'app-chip app-chip-entry-points',
      dense: true
    }),
    content: row => row.type
  },
  {
    name: 'servers',
    align: 'right',
    label: 'Servers',
    sortable: true,
    fieldToProps: () => ({ class: 'servers-label' }),
    content: function (value) {
      if (value.loadBalancer && value.loadBalancer.servers) {
        return value.loadBalancer.servers.length
      }
      return 0
    }
  },
  {
    name: 'service',
    align: 'left',
    label: 'Service',
    component: QChip,
    sortable: true,
    fieldToProps: () => ({ class: 'app-chip app-chip-service', dense: true }),
    content: row => row.service
  },
  {
    name: 'provider',
    align: 'center',
    label: 'Provider',
    sortable: true,
    fieldToProps: row => ({ name: row.provider }),
    component: ProviderIcon
  },
  {
    name: 'priority',
    align: 'left',
    label: 'Priority',
    sortable: true,
    component: QChip,
    fieldToProps: () => ({ class: 'app-chip app-chip-accent', dense: true }),
    content: row => {
      return {
        short: String(row.priority).length > 10 ? String(row.priority).substring(0, 10) + '...' : row.priority,
        long: row.priority
      }
    }
  }
]

const columnsByResource = {
  routers: [
    'status',
    'rule',
    'entryPoints',
    'name',
    'service',
    'tls',
    'provider',
    'priority'
  ],
  udpRouters: ['status', 'entryPoints', 'name', 'service', 'provider'],
  services: ['status', 'name', 'type', 'servers', 'provider'],
  middlewares: ['status', 'name', 'type', 'provider']
}

const propsByType = {
  'http-routers': {
    columns: columnsByResource.routers
  },
  'tcp-routers': {
    columns: columnsByResource.routers
  },
  'udp-routers': {
    columns: columnsByResource.udpRouters
  },
  'http-services': {
    columns: columnsByResource.services
  },
  'tcp-services': {
    columns: columnsByResource.services
  },
  'udp-services': {
    columns: columnsByResource.services
  },
  'http-middlewares': {
    columns: columnsByResource.middlewares
  },
  'tcp-middlewares': {
    columns: columnsByResource.middlewares
  }
}

const GetTablePropsMixin = {
  methods: {
    getTableProps ({ type }) {
      return {
        onRowClick: row =>
          this.$router.push({
            path: `/${type.replace('-', '/', 'gi')}/${encodeURIComponent(row.name)}`
          }),
        columns: allColumns.filter(c =>
          getProperty(propsByType, `${type}.columns`, []).includes(c.name)
        )
      }
    }
  }
}

export default GetTablePropsMixin
