<template>
  <div>
    <q-chip
      v-for="(chip, index) in list"
      :key="index"
      :dense="dense"
      :class="classNames"
    >
      {{ chip }}
    </q-chip>
  </div>
</template>

<script>
export default {
  props: {
    dense: Boolean,
    classNames: Array[String],
    list: Array[Object]
  }
}
</script>
