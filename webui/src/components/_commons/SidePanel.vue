<template>
  <div class="panel">
    <div
      class="panel-backdrop"
      @click="close"
      v-if="isOpen"
    ></div>
    <transition name="slide">
      <div v-if="isOpen" class="panel-content">
        <slot></slot>
      </div>
    </transition>
  </div>
</template>
<script>

export default {
  props: ['isOpen'],
  methods: {
    close () {
      this.$emit('onClose')
    }
  }
}
</script>

<style scoped lang="scss">
  @import "../../css/sass/mixins";

  .slide-enter-active,
  .slide-leave-active {
    transition: transform 0.2s ease;
  }

  .slide-enter,
  .slide-leave-to {
    transform: translateX(100%);
    transition: all 150ms ease-in 0s;
  }

  .panel-backdrop {
    z-index: 3000;
    background-color: rgba(255, 255, 255, 0.47);
    width: 100vw;
    height: 100vh;
    position: fixed;
    top: 0;
    left: 0;
    cursor: pointer;
  }

  .panel-content {
    z-index: 9999;
    overflow-y: auto;
    background-color: white;
    position: fixed;
    right: 0;
    top: 0;
    height: 100vh;
    padding: 0;
    width: 100vw;
    border-top-left-radius: 20px;
    border-bottom-left-radius: 20px;
    box-shadow: 2px 0 6px 0 #000;

    @include respond-to(md) {
      width: 80vw;
      max-width: 1500px;
    }
  }
</style>
