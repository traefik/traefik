<template>
  <q-card flat bordered>
    <q-card-section>
      <div class="row items-center no-wrap">
        <div class="col">
          <div class="text-h6 text-weight-bold">{{getName}}</div>
        </div>
        <div class="col-auto">
          <q-btn :to="getUrl" color="accent" dense flat icon-right="eva-arrow-forward-outline" no-caps label="Explore" size="md" class="text-weight-bold"/>
        </div>
      </div>
    </q-card-section>
    <q-card-section>
      <div class="row items-center q-col-gutter-md">
        <div class="col-12 col-sm-6">
          <ChartDoughnut
            :chartdata="getChartdata()"
            :options="options"/>
        </div>
        <div class="col-12 col-sm-6">
          <q-list>
            <q-item class="label-state">
              <q-item-section avatar>
                <avatar-state state="positive"/>
              </q-item-section>
              <q-item-section class="label-state-text">
                <q-item-label>Success</q-item-label>
                <q-item-label caption lines="1">{{getSuccess(true)}}%</q-item-label>
              </q-item-section>
              <q-item-section side class="label-state-side">
                {{getSuccess()}}
              </q-item-section>
            </q-item>
            <q-item class="label-state">
              <q-item-section avatar>
                <avatar-state state="warning"/>
              </q-item-section>
              <q-item-section class="label-state-text">
                <q-item-label>Warnings</q-item-label>
                <q-item-label caption lines="1">{{getWarnings(true)}}%</q-item-label>
              </q-item-section>
              <q-item-section side class="label-state-side">
                {{getWarnings()}}
              </q-item-section>
            </q-item>
            <q-item class="label-state">
              <q-item-section avatar>
                <avatar-state state="negative"/>
              </q-item-section>
              <q-item-section class="label-state-text">
                <q-item-label>Errors</q-item-label>
                <q-item-label caption lines="1">{{getErrors(true)}}%</q-item-label>
              </q-item-section>
              <q-item-section side class="label-state-side">
                {{getErrors()}}
              </q-item-section>
            </q-item>
          </q-list>
        </div>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
import Helps from '../../_helpers/Helps'
import ChartDoughnut from '../_commons/ChartDoughnut'
import AvatarState from '../_commons/AvatarState'

export default {
  name: 'PanelChart',
  props: ['name', 'data', 'type'],
  components: {
    ChartDoughnut,
    AvatarState
  },
  data () {
    return {
      loading: true,
      options: {
        legend: {
          display: false
        },
        animation: {
          duration: 1000
        },
        tooltips: {
          enabled: true
        }
      }
    }
  },
  computed: {
    getName () {
      return Helps.capFirstLetter(this.name)
    },
    getUrl () {
      return `/${this.type}/${this.getName.toLowerCase()}`
    }
  },
  methods: {
    getSuccess (inPercent = false) {
      const num = this.data.total - (this.data.errors + this.data.warnings)
      let result = 0
      if (inPercent) {
        result = Helps.getPercent(num, this.data.total).toFixed(0)
      } else {
        result = num
      }
      return isNaN(result) || result < 0 ? 0 : result
    },
    getWarnings (inPercent = false) {
      const num = this.data.warnings
      let result = 0
      if (inPercent) {
        result = Helps.getPercent(num, this.data.total).toFixed(0)
      } else {
        result = num
      }
      return isNaN(result) || result < 0 ? 0 : result
    },
    getErrors (inPercent = false) {
      const num = this.data.errors
      let result = 0
      if (inPercent) {
        result = Helps.getPercent(num, this.data.total).toFixed(0)
      } else {
        result = num
      }
      return isNaN(result) || result < 0 ? 0 : result
    },
    getData () {
      return [this.getSuccess(), this.getWarnings(), this.getErrors()]
    },
    getChartdata () {
      if (this.getData()[0] === 0 && this.getData()[1] === 0 && this.getData()[2] === 0) {
        this.options.tooltips.enabled = false
        return {
          datasets: [{
            backgroundColor: [
              this.$q.dark.isActive ? '#2d2d2d' : '#f2f3f5'
            ],
            borderColor: [
              this.$q.dark.isActive ? '#1d1d1d' : '#fff'
            ],
            data: [1]
          }]
        }
      } else {
        this.options.tooltips.enabled = true
        return {
          datasets: [{
            backgroundColor: [
              '#00a697',
              '#db7d11',
              '#ff0039'
            ],
            borderColor: [
              this.$q.dark.isActive ? '#1d1d1d' : '#fff',
              this.$q.dark.isActive ? '#1d1d1d' : '#fff',
              this.$q.dark.isActive ? '#1d1d1d' : '#fff'
            ],
            data: this.getData()
          }],
          labels: [
            'Success',
            'Warnings',
            'Errors'
          ]
        }
      }
    }
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .label-state {
    min-height: 32px;
    padding: 8px;
    .q-item__section--avatar{
      min-width: 32px;
      padding: 0 8px 0 0;
    }
    &-text{
      .q-item__label{
        font-size: 16px;
        line-height: 16px !important;
        font-weight: 600;
      }
      .q-item__label--caption{
        font-size: 14px;
        line-height: 14px !important;
        font-weight: 500;
        color: $app-text-grey;
      }
    }
    &-side{
      font-size: 22px;
      font-weight: 700;
      padding: 0 0 0 8px;
      color: inherit;
    }
  }
</style>
