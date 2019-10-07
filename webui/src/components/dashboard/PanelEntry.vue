<template>
  <q-card flat bordered v-bind:class="['panel-entry', {'panel-entry-detail':type === 'detail'}, {'panel-entry-focus':focus}, {'panel-entry-ex-size':exSize}]">
    <q-card-section>
      <div class="row items-center no-wrap">
        <div class="col">
          <div class="text-subtitle2">{{name}}</div>
        </div>
      </div>
    </q-card-section>
    <q-card-section>
      <div class="text-h3 text-center text-weight-bold ellipsis">
        <span>{{address}}</span>
        <q-tooltip>{{address}}</q-tooltip>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
export default {
  name: 'PanelEntry',
  props: ['address', 'name', 'type', 'focus', 'exSize']
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .panel-entry {
    .text-subtitle2 {
      font-weight: 600;
      letter-spacing: 3px;
      color: $app-text-grey;
      text-transform: uppercase;
      text-align: center;
    }
    &-detail{
      .text-subtitle2 {
        font-size: 11px;
        font-weight: 600;
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
    &-ex-size{
      .text-h3 {
        font-size: 22px;
      }
    }
  }
</style>
