<template>
  <div class="fixed-center text-center q-pa-md">
    <h1 class="q-ma-md"><strong>404</strong></h1>
    <h5 class="q-ma-md">I'm sorry, nothing around here ...</h5>
    <q-btn
      color="secondary"
      style="width:200px;"
      @click="$router.push('/')"
    >Go back</q-btn>
  </div>
</template>

<script>
export default {
  name: 'Error404'
}
</script>
