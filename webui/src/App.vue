<template>
  <div id="app">
    <nav class="navbar" role="navigation" aria-label="main navigation">
      <div class="navbar-brand">
        <a class="navbar-item" href="/dashboard">
          <img
            src="./assets/images/traefik_logo@3x.svg"
            alt="Traefik Webui"
            width="112"
            height="28"
          />
        </a>

        <a
          :class="{ 'is-active': isActive }"
          role="button"
          class="navbar-burger burger"
          aria-label="menu"
          aria-expanded="false"
          data-target="navbarBasicExample"
          @click="toggle"
        >
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
        </a>
      </div>

      <div
        :class="{ 'is-active': isActive }"
        class="navbar-menu"
        v-if="version.Version"
      >
        <div class="navbar-start">
          <a class="navbar-item" :href="documentationUrl">
            Documentation
          </a>
        </div>

        <div class="navbar-end">
          <div class="navbar-item">Version: {{ version.Version }}</div>
        </div>
      </div>
    </nav>
    <router-view />
  </div>
</template>

<script>
export default {
  name: "App",
  data: () => ({
    version: {},
    isActive: false
  }),
  computed: {
    parsedVersion() {
      if (this.version.Version === "dev") {
        return "master";
      } else {
        const matches = this.version.Version.match(/^(v?\d+\.\d+)/);
        return matches ? matches[1] : null;
      }
    },
    documentationUrl() {
      return `https://docs.traefik.io/${this.parsedVersion}`;
    }
  },
  created() {
    this.fetchVersion();
  },
  methods: {
    fetchVersion() {
      return fetch("/api/version")
        .then(response => response.json())
        .then(response => (this.version = response));
    },
    toggle() {
      this.isActive = !this.isActive;
    }
  }
};
</script>

<style lang="sass">

@import 'styles/typography'
@import 'styles/colors'

html
  font-family: $open-sans
  height: 100%
  background: $background
</style>
