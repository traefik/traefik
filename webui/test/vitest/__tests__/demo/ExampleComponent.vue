<template>
  <div>
    <p>{{ title }}</p>
    <q-list>
      <q-item v-for="todo in todos" :key="todo.id" clickable @click="increment">
        {{ todo.id }} - {{ todo.content }}
      </q-item>
    </q-list>

    <p>Count: {{ todoCount }} / {{ totalCount }}</p>
    <p>Active: {{ active ? 'yes' : 'no' }}</p>
    <p>Clicks on todos: {{ clickCount }}</p>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  todos: {
    type: Array,
    default() {
      return [];
    },
  },
  totalCount: {
    type: Number,
    required: true,
  },
  active: Boolean,
});

const clickCount = ref(0);
function increment() {
  clickCount.value += 1;
  return clickCount.value;
}

const todoCount = computed(() => props.todos.length);
</script>
