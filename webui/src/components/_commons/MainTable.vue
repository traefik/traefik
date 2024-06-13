<template>
  <div class="table-wrapper">
    <q-infinite-scroll
      ref="scroller"
      :offset="250"
      @load="handleLoadMore"
    >
      <q-markup-table>
        <thead>
          <tr class="table-header">
            <th
              v-for="column in columns"
              :key="column.name"
              :class="`text-${column.align}`"
            >
              {{ column.label }}
            </th>
          </tr>
        </thead>
        <tfoot v-if="!data || !data.length">
          <tr>
            <td colspan="100%">
              <q-icon
                name="warning"
                style="font-size: 1.5rem"
              /> No data available
            </td>
          </tr>
        </tfoot>
        <tbody>
          <tr
            v-for="row in data"
            :key="row.name"
            class="cursor-pointer"
            @click="onRowClick(row)"
          >
            <template v-for="column in columns">
              <td
                v-if="getColumn(column.name).component"
                :key="column.name"
                :class="`text-${getColumn(column.name).align}`"
              >
                <component
                  :is="getColumn(column.name).component"
                  v-bind="getColumn(column.name).fieldToProps(row)"
                >
                  <template v-if="getColumn(column.name).content">
                    {{ getColumn(column.name).content(row) }}
                  </template>
                </component>
              </td>
              <td
                v-if="!getColumn(column.name).component"
                :key="column.name"
                :class="`text-${getColumn(column.name).align}`"
                v-bind="getColumn(column.name).fieldToProps(row)"
              >
                <span>
                  {{ getColumn(column.name).content ? getColumn(column.name).content(row) : row[column.name] }}
                </span>
              </td>
            </template>
          </tr>
        </tbody>
      </q-markup-table>
      <template
        v-if="loading"
        #loading
      >
        <div class="row justify-center q-my-md">
          <q-spinner-dots
            color="app-grey"
            size="40px"
          />
        </div>
      </template>
    </q-infinite-scroll>
    <q-page-scroller
      v-if="endReached"
      position="bottom"
      :scroll-offset="150"
      class="back-to-top"
    >
      <q-btn
        color="primary"
        small
      >
        Back to top
      </q-btn>
    </q-page-scroller>
  </div>
</template>

<script>
import { defineComponent } from 'vue'
import { QMarkupTable, QInfiniteScroll, QSpinnerDots, QPageScroller } from 'quasar'

export default defineComponent({
  name: 'MainTable',
  components: {
    QMarkupTable,
    QInfiniteScroll,
    QSpinnerDots,
    QPageScroller
  },
  props: {
    data: { type: Object, default: undefined, required: false },
    columns: Array[Object],
    loading: Boolean,
    onLoadMore: { type: Function, default: undefined, required: false },
    endReached: Boolean,
    onRowClick: { type: Function, default: undefined, required: false }
  },
  methods: {
    getColumn (columnName) {
      return this.columns.find(c => c.name === columnName) || {}
    },
    handleLoadMore (index, done) {
      if (!this?.onLoadMore) {
        done()
      } else {
        this.onLoadMore({ page: index })
          .then(() => done())
          .catch(() => done(true))
      }
    }
  }
})
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .table-wrapper {
    :deep(.q-table__container) {
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

  .servers-label {
    font-size: 14px;
    font-weight: 600;
  }
</style>
