<template>
  <span class="badge" :class="badgeClass">{{ name }}</span>
</template>

<script setup lang="ts">
const props = defineProps<{
  name: string,
  badgeClass: string,
}>();
</script>