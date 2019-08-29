<template>
  <q-card flat bordered v-bind:class="['panel-entry', {'panel-entry-detail':type === 'detail'}, {'panel-entry-focus':focus}]">
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
  props: ['address', 'name', 'type', 'focus']
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .panel-entry {
    &-detail{
      .text-subtitle2 {
        font-size: 11px;
        line-height: 11px;
        text-align: left;
      }
      .text-h3 {
        font-size: 16px;
        text-align: left;
        line-height: 16px;
      }
    }
    &-focus {
      border: solid 2px $accent;
    }
  }
</style>
