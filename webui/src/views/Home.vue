<template>
  <main class="home section">
    <section class="container" v-if="entrypoints">
      <h1 class="title is-4">Entrypoints</h1>
      <nav class="level">
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
    </section>
    <section class="container">
      <div class="columns">
        <section v-if="overview.http" class="column is-6">
          <h1 class="title is-4">HTTP</h1>
          <div class="columns">
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
        <section v-if="overview.tcp" class="column is-6">
          <h1 class="title is-4">TCP</h1>

          <div class="columns">
            <div class="column is-4">
              <canvas id="tcp-routers" />
            </div>
            <div class="column is-4">
              <canvas id="tcp-services" />
            </div>
          </div>
        </section>
      </div>
    </section>

    <section class="container">
      <h1 class="title is-4">Features</h1>
      <div class="tile is-ancestor">
        <div
          class="tile is-parent"
          v-for="(feature, key) of overview.features"
          :key="key"
        >
          <div
            class="tile is-child notification"
            :class="{ 'is-success': feature }"
          >
            <p class="title">{{ key }}</p>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<script>
import Chart from "chart.js";

export default {
  name: "home",
  data: () => ({
    entrypoints: [],
    overview: {},
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
      entity = { errors: 0, warnings: 0, total: 0 },
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
                "Middleware"
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
  margin-bottom: 1rem;
}
</style>
