<template>
  <main class="home section">
    <section class="container hero is-info">
      <div class="hero-body">
        <div class="container">
          <h1 class="title">
            🚧 Work in progress...
          </h1>
          <h2 class="subtitle">
            <p>
              In the meantime, you can review your current configuration by
              using the
              <a href="/api/rawdata"
                >/api/rawdata <i class="fas fa-external-link-alt"
              /></a>
              endpoint.
            </p>
            <p>
              Also, please keep your <i class="fa fa-eye" /> on our
              <a href="https://docs.traefik.io/v2.0/operations/dashboard/"
                >documentation</a
              >
              to stay informed
            </p>
          </h2>
        </div>
      </div>
    </section>

    <section class="container" v-if="entrypoints.length">
      <p class="title is-4">Entrypoints</p>
      <div class="tile is-child box columns">
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
          <EntityStateDoughnut
            :entity="overview.http.routers"
            title="Routers"
          />
        </div>
        <div class="column is-4">
          <EntityStateDoughnut
            :entity="overview.http.middlewares"
            title="Middlewares"
          />
        </div>
        <div class="column is-4">
          <EntityStateDoughnut
            :entity="overview.http.services"
            title="Services"
          />
        </div>
      </div>
    </section>

    <section class="container" v-if="overview.tcp">
      <p class="title is-4">TCP</p>
      <div class="tile is-child box columns">
        <div class="column is-4">
          <EntityStateDoughnut :entity="overview.tcp.routers" title="Routers" />
        </div>
        <div class="column is-4">
          <EntityStateDoughnut
            :entity="overview.tcp.services"
            title="Services"
          />
        </div>
      </div>
    </section>

    <section class="container">
      <p class="title is-4">Features</p>
      <div class="tile is-child box columns">
        <div class="tile is-ancestor">
          <FeatureTile
            v-for="(feature, key) of overview.features"
            :key="key"
            :title="key"
            :modifier="{ 'is-success': feature, 'is-danger': !feature }"
          />
        </div>
      </div>
    </section>
  </main>
</template>

<script>
import FeatureTile from "../components/Tile";
import EntityStateDoughnut from "../components/EntityStateDoughnut";

export default {
  name: "home",
  components: {
    FeatureTile,
    EntityStateDoughnut
  },
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
    fetchOverview() {
      return fetch("/api/overview")
        .then(response => response.json())
        .then(response => (this.overview = response));
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
