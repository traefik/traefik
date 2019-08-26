<template>
  <q-card flat bordered>
    <q-card-section>
      <div class="row items-center no-wrap">
        <div class="col">
          <div class="text-subtitle2 text-uppercase text-center text-app-grey" style="letter-spacing: 3px;">{{name}}</div>
        </div>
      </div>
    </q-card-section>
    <q-card-section>
      <div class="text-h3 text-center text-weight-bold">{{address}}</div>
    </q-card-section>
  </q-card>
</template>

<script>
export default {
  name: 'PanelEntry',
  props: ['address', 'name']
}
</script>

<style scoped>

</style>
