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
      </div>
    </nav>
    <router-view />
  </div>
</template>

<style lang="sass">

@import 'styles/typography'
@import 'styles/colors'

html
  font-family: $open-sans
  height: 100%
  background: $background
</style>
