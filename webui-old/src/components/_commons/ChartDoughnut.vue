<script>
import { Doughnut } from 'vue-chartjs'
import isEqual from 'lodash.isequal'

export default {
  extends: Doughnut,
  props: {
    chartdata: {
      type: Object,
      default: null
    },
    options: {
      type: Object,
      default: null
    }
  },
  watch: {
    chartdata: function (newData, oldData) {
      // TODO - bug, 'update()' not update the chart, replace for renderChart()
      // console.log('new data from watcher...', newData, oldData, isEqual(newData.datasets[0].data, oldData.datasets[0].data))
      if (!isEqual(newData.datasets[0].data, oldData.datasets[0].data)) {
        // this.$data._chart.update()
        this.renderChart(this.chartdata, this.options)
      }
    },
    '$q.dark.isActive' (val) {
      this.renderChart(this.chartdata, this.options)
    }
  },
  mounted () {
    this.renderChart(this.chartdata, this.options)
  }
}
</script>
