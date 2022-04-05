<template>
  <BaseTable :data="data" :headers="headers"></BaseTable>
</template>

<script setup lang="ts">
import BaseTable from "../components/BaseTable.vue";

type TabDataItem = {
  id: string;
  first: string;
  last: string;
}

const data: TabDataItem[] = [
  {
    id: "primo",
    first: "First",
    last: "Amangi"
  },
  {
    id: "segundo",
    first: "second",
    last: "city"
  },
  {
    id: "tri",
    first: "cera",
    last: "tops"
  }
];

const headers = ["ID", "FIRST", "LAST"];

</script>