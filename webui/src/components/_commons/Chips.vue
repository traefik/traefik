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
import { defineComponent } from 'vue'

export default defineComponent({
  props: ['dense', 'classNames', 'list']
})
</script>
