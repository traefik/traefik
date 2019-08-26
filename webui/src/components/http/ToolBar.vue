<template>
  <q-toolbar class="row no-wrap items-center">
    <q-tabs align="left" inline-label indicator-color="transparent" stretch>
      <q-route-tab :to="`/http/routers`" no-caps label="HTTP Routers">
        <q-badge align="middle" label="10" class="q-ml-sm"/>
      </q-route-tab>
      <q-route-tab :to="`/http/services`" no-caps label="HTTP Services">
        <q-badge align="middle" label="12" class="q-ml-sm"/>
      </q-route-tab>
      <q-route-tab :to="`/http/middlewares`" no-caps label="HTTP Middlewares">
        <q-badge align="middle" label="8" class="q-ml-sm"/>
      </q-route-tab>
    </q-tabs>
  </q-toolbar>
</template>

<script>
export default {
  name: 'HTTPToolBar',
  data () {
    return {
    }
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .q-toolbar {
    min-height: 48px;
    padding: 0;
  }

  .q-tabs {
    /deep/ .q-tabs__content {
      .q-tab__label {
        color: $app-text-grey;
        font-size: 16px;
        font-weight: 700;
      }
      .q-badge {
        font-size: 13px;
        font-weight: 700;
        border-radius: 12px;
        text-align: center;
        align-items: center;
        justify-content: center;
        min-width: 30px;
        padding: 6px;
        color: $app-text-grey;
        background-color: rgba( $app-text-grey, .1 );
      }
      .q-tab--active {
        .q-tab__label {
          color: $accent;
        }
        .q-badge {
          color: $accent;
          background-color: rgba( $accent, .1 );
        }
      }
    }
  }
</style>
