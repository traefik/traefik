<template>
  <canvas />
</template>

<script>
import Chart from "chart.js";

Chart.plugins.register({
  afterDraw: function(chart) {
    if (chart.data.datasets[0].data.reduce((acc, it) => acc + it, 0) === 0) {
      var ctx = chart.chart.ctx;
      var width = chart.chart.width;
      var height = chart.chart.height;
      chart.clear();

      ctx.save();
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      ctx.font = "16px normal 'Helvetica Nueue'";
      ctx.fillText(`No ${chart.options.title.text}`, width / 2, height / 2);
      ctx.restore();
    }
  }
});

export default {
  name: "EntityStateDoughnut",
  props: {
    entity: {
      type: Object,
      default: () => ({
        errors: 0,
        warnings: 0,
        total: 0
      })
    },
    title: {
      type: String,
      required: true
    }
  },
  computed: {
    data() {
      return [
        this.entity.errors,
        this.entity.warnings,
        this.entity.total - (this.entity.errors + this.entity.warnings)
      ];
    }
  },
  mounted() {
    new Chart(this.$el, {
      type: "doughnut",
      data: {
        datasets: [
          {
            data: this.data,
            backgroundColor: [
              "hsl(348, 100%, 61%)",
              "hsl(48, 100%, 67%)",
              "hsl(141, 71%, 48%)"
            ]
          }
        ],
        labels: ["errors", "warnings", "success"]
      },
      options: {
        title: {
          display: true,
          text: this.title
        },
        legend: {
          display: false
        }
      }
    });
  }
};
</script>
