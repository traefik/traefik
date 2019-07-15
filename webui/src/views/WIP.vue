<template>
  <main class="wip">
    <img src="../assets/images/traefik.logo.svg" alt="logo" />
    <header>
      <h1 class="title">
        <i class="fa fa-exclamation-triangle"></i>
        Work in progress...
      </h1>
      <p>
        In the meantime, you can review your current configuration by using the
        <a href="/api/rawdata">/api/rawdata</a> endpoint <br /><br />
        Also, please keep your <i class="fa fa-eye"></i> on our
        <a href="https://docs.traefik.io/v2.0/operations/dashboard/"
          >documentation</a
        >
        to stay informed
      </p>
      <p></p>
    </header>
  </main>
</template>

<script>
export default {
  name: "home"
};
</script>

<style lang="sass">
.wip
  display: flex
  flex-direction: column
  align-items: center
  justify-content: center
  height: 100vh

  .title
    font-size: 4em
</style>
