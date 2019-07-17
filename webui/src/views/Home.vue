<template>
  <main class="home section">
    <section class="container panel">
      <p class="panel-heading ">🚧 Work in progress...</p>
      <div class="panel-block">
        <div>
          <p>
            In the meantime, you can review your current configuration by using
            the <a href="/api/rawdata">/api/rawdata</a> endpoint.
          </p>
          <p>
            Also, please keep your <i class="fa fa-eye" /> on our
            <a href="https://docs.traefik.io/v2.0/operations/dashboard/"
              >documentation</a
            >
            to stay informed
          </p>
        </div>
      </div>
    </section>

    <section class="container panel" v-if="entrypoints.length">
      <p class="panel-heading ">Entrypoints</p>
      <div class="panel-block">
        <nav class="level" :style="{ flex: '1 1' }">
          <div
            class="level-item has-text-centered"
            v-for="entrypoint in entrypoints"
            :key="entrypoint.name"
          >
            <div>
              <p class="heading">{{ entrypoint.name }}</p>
              <p class="title">{{ entrypoint.address }}</p>
            </div>
          </div>
        </nav>
      </div>
    </section>

    <section class="container" v-if="overview.http">
      <p class="title is-4">HTTP</p>
      <div class="tile is-child box columns">
        <div class="column is-4">
          <canvas id="http-routers" />
        </div>
        <div class="column is-4">
          <canvas id="http-middlewares" />
        </div>
        <div class="column is-4">
          <canvas id="http-services" />
        </div>
      </div>
    </section>

    <section class="container" v-if="overview.tcp">
      <p class="title is-4">TCP</p>
      <div class="tile is-child box columns">
        <div class="column is-6">
          <canvas id="tcp-routers" />
        </div>
        <div class="column is-6">
          <canvas id="tcp-services" />
        </div>
      </div>
    </section>

    <section class="container panel">
      <p class="panel-heading">Features</p>
      <div class="panel-block">
        <div class="tile is-ancestor">
          <div
            class="tile is-parent"
            v-for="(feature, key) of overview.features"
            :key="key"
          >
            <div
              class="tile is-child notification"
              :class="{ 'is-success': feature, 'is-danger': !feature }"
            >
              <p class="title">{{ key }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<script>
import Chart from "chart.js";

Chart.plugins.register({
  afterDraw: function(chart) {
    if (chart.data.datasets[0].data.reduce((acc, it) => acc + it, 0) === 0) {
      var ctx = chart.chart.ctx;
      var width = chart.chart.width;
      var height = chart.chart.height
      chart.clear();

      ctx.save();
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.font = "16px normal 'Helvetica Nueue'";
      ctx.fillText(`No ${chart.options.title.text}`, width / 2, height / 2);
      ctx.restore();
    }
  }
});

export default {
  name: "home",
  data: () => ({
    entrypoints: [],
    overview: {
      features: []
    },
    charts: {
      http: {
        routers: null,
        middlewares: null,
        services: null
      },
      tcp: {
        routers: null,
        services: null
      }
    },
    interval: null
  }),
  methods: {
    buildDoughnutChart(
      selector,
      entity = { errors: 2, warnings: 2, total: 6 },
      name
    ) {
      return new Chart(this.$el.querySelector(selector), {
        type: "doughnut",
        data: {
          datasets: [
            {
              data: [
                entity.errors,
                entity.warnings,
                entity.total - (entity.errors + entity.warnings)
              ],
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
            text: name
          },
          legend: {
            display: false
          }
        }
      });
    },
    fetchOverview() {
      return fetch("/api/overview")
        .then(response => response.json())
        .then(response => (this.overview = response))
        .then(() => {
          this.charts = {
            http: {
              routers: this.buildDoughnutChart(
                "#http-routers",
                this.overview.http.routers,
                "Routers"
              ),
              middlewares: this.buildDoughnutChart(
                "#http-middlewares",
                this.overview.http.middlewares,
                "Middlewares"
              ),
              services: this.buildDoughnutChart(
                "#http-services",
                this.overview.http.services,
                "Services"
              )
            },
            tcp: {
              routers: this.buildDoughnutChart(
                "#tcp-routers",
                this.overview.tcp.routers,
                "Routers"
              ),
              services: this.buildDoughnutChart(
                "#tcp-services",
                this.overview.tcp.services,
                "Services"
              )
            }
          };
        });
    },
    fetchEntrypoints() {
      return fetch("/api/entrypoints")
        .then(response => response.json())
        .then(response => (this.entrypoints = response));
    }
  },
  async mounted() {
    await this.fetchOverview();
    await this.fetchEntrypoints();
    this.interval = setInterval(() => {
      this.fetchOverview();
      this.fetchEntrypoints();
    }, 60000);
  },
  beforeDestroy() {
    clearInterval(this.interval);
  }
};
</script>

<style lang="scss">
.home section {
  margin-bottom: 1.5rem;
}
</style>
