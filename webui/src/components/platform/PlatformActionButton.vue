<template>
  <q-btn color="accent" class="btn" @click="$emit('click', $event)" :bind="$attrs">{{label}}</q-btn>
</template>

<script>

export default {
  name: 'ConnectButton',
  props: ['label']
}
</script>

<style>
  .btn {
    font-family: 'Nunito', 'Roboto', sans-serif;
    font-weight: bold;
    font-size: 14px;
    text-transform: inherit;
  }
</style>
