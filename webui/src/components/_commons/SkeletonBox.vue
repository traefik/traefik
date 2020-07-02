<template>
  <span
    :style="{ height, width: computedWidth }"
    v-bind:class="['SkeletonBox']"
  />
</template>

<script>
export default {
  name: `SkeletonBox`,
  props: {
    maxWidth: {
      default: 100,
      type: Number
    },
    minWidth: {
      default: 80,
      type: Number
    },
    height: {
      default: `2em`,
      type: String
    },
    width: {
      default: null,
      type: String
    }
  },
  computed: {
    computedWidth () {
      return this.width || `${Math.floor((Math.random() * (this.maxWidth - this.minWidth)) + this.minWidth)}%`
    }
  }
}
</script>

<style scoped lang="scss">
  .SkeletonBox {
    display: inline-block;
    position: relative;
    vertical-align: middle;
    overflow: hidden;
    background-color: #E0E0E0;
    border-radius: 4px;
    &.dark{
      background-color: #9E9E9E;
    }
    &::after {
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
      will-change: transform;
      transform: translateX(-100%) translateZ(0);
      background-image: linear-gradient(
          90deg,
          rgba(#fff, 0) 0,
          rgba(#fff, 0.2) 20%,
          rgba(#fff, 0.5) 60%,
          rgba(#fff, 0)
      );
      // TODO - fix high cpu usage
      // animation: shimmer 5s infinite;
      content: '';
    }
    @keyframes shimmer {
      from { transform: translateX(-100%) translateZ(0); }
      to { transform: translateX(100%) translateZ(0); }
    }
  }

  .body--dark .SkeletonBox {
    background-color: #525252;

    &.dark {
      background-color: #333;
    }

    &::after {
      background-image: linear-gradient(
          90deg,
          rgba(#5e5e5e, 0) 0,
          rgba(#5e5e5e, 0.2) 20%,
          rgba(#5e5e5e, 0.5) 60%,
          rgba(#5e5e5e, 0)
      );
    }
  }
</style>
