<template>
  <div class="tile is-parent">
    <div class="tile is-child notification" :class="modifier">
      <p class="title">{{ title }}</p>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    modifier: {
      type: Object,
      default: () => ({})
    },
    title: {
      type: String,
      required: true
    }
  }
};
</script>
