<template>
  <nav aria-label="Page navigation example">
    <ul class="pagination">
      <li class="page-item">
        <a class="page-link" href="#" aria-label="Previous">
          <span aria-hidden="true">&laquo;</span>
        </a>
      </li>
      <li class="page-item">
        <a class="page-link" href="#">1</a>
      </li>
      <li class="page-item">
        <a class="page-link" href="#">2</a>
      </li>
      <li class="page-item">
        <a class="page-link" href="#">3</a>
      </li>
      <li class="page-item">
        <a class="page-link" href="#" aria-label="Next">
          <span aria-hidden="true">&raquo;</span>
        </a>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const props = withDefaults(defineProps<{
  currentPage: number,
  totalRows: number,
  lastPage: number,
  numRows?: number,
  numTabs?: number,
}>(), {
  numRows: 20,
  numTabs: 5,
});

const firstTab = ref(1);

// const lastTab = computed(() => {
//   return Math.min(props.lastPage + 1, firstTab.value + props.numTabs);
// });

// const prevDisabled = computed(() => {
//   return firstTab.value < 1;
// });

// const nextDisabled = computed(() => {
//   return lastTab.value >= props.lastPage;
// });

function fillPagination(currPage: number, lastPage: number) {
  const maxTabs = props.numTabs;

  let lastTab = Math.min(lastPage + 1, firstTab.value + maxTabs);
}


</script>