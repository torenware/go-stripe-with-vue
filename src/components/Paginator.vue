<template>
  <nav v-if="lastPage > 1" aria-label="Page navigation example">
    <ul class="pagination">
      <li class="page-item">
        <a class="page-link" href="#" :class="disabledClass(prevDisabled)" aria-label="Previous">
          <span aria-hidden="true">&laquo;</span>
        </a>
      </li>
      <li v-for="tab in tabData" class="page-item" :key="tab.page">
        <a
          class="page-link"
          href="#"
          :class="{ 'bg-info': tab.page === currentPage }"
          @click="changePage(tab.page)"
        >{{ tab.page }}</a>
      </li>
      <li class="page-item">
        <a class="page-link" href="#" :class="disabledClass(nextDisabled)" aria-label="Next">
          <span aria-hidden="true">&raquo;</span>
        </a>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';

const props = withDefaults(defineProps<{
  currentPage: number,
  totalRows: number,
  lastPage: number,
  numRows?: number,
  numTabs?: number,
}>(), {
  currentPage: 0,
  totalRows: 0,
  lastPage: 0,
  numRows: 20,
  numTabs: 5,
});

const emit = defineEmits<{
  (event: "page-change", currentPage: number): void
}>();

const changePage = (page: number) => {
  if (page !== props.currentPage) {
    console.log("change page to", page);
    emit("page-change", page);
  }
}

const firstTab = ref(1);

type PageTags = {
  page: number;
}

const tabData = computed(() => {
  const tabs: PageTags[] = [];
  for (let ndx = firstTab.value; ndx <= lastTab.value; ndx++) {
    tabs.push({
      page: ndx
    })
  }
  return tabs;
});

const disabledClass = (state: boolean) => {
  return {
    disabled: state
  }
};

onMounted(() => {
  console.log("props:", props);
});

const lastTab = computed(() => {
  return Math.min(props.lastPage, firstTab.value + props.numTabs);
});

const prevDisabled = computed(() => {
  return firstTab.value <= 1;
});

const nextDisabled = computed(() => {
  console.log("disabled", lastTab.value, props.lastPage);
  return lastTab.value >= props.lastPage;
});


</script>

<style scoped>
a.disabled {
  pointer-events: none;
  opacity: 0.6;
  color: lightgray;
}

/**
 Make the "ring around the window" disappear
 https://stackoverflow.com/a/69584423/8600734
 */
.page-link:focus {
  box-shadow: 0 0 0 0;
}
</style>