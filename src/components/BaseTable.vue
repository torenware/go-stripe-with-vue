<template>
  <section class="d-flex flex-column align-items-center">
    <table class="table">
      <thead>
        <tr>
          <slot name="header-row">
            <td v-for="cell in headers" :key="cell">{{ cell }}</td>
          </slot>
        </tr>
      </thead>
      <tbody>
        <slot name="body-rows">
          <tr v-for="item in data" :key="item.id">
            <td>{{ item.id }}</td>
            <td>{{ item.first }}</td>
            <td>{{ item.last }}</td>
          </tr>
        </slot>
      </tbody>
    </table>
    <Paginator
      @page-change="pageChange"
      :current-page="currentPage"
      :total-rows="totalRows"
      :last-page="lastPage"
    ></Paginator>
  </section>
</template>

<script>
import Paginator from "../components/Paginator.vue";
import { onMounted } from "vue";

export default {
  emits: ['page-change'],
  props: {
    data: Array,
    headers: Array,
    currentPage: {
      type: Number,
      required: false,
      default: 0,
    },
    totalRows: {
      type: Number,
      required: false,
      default: 0,
    },
    lastPage: {
      type: Number,
      required: false,
      default: 0,
    },

  },
  setup(props, ctx) {

    function pageChange(page) {
      console.log("got event in BT", page);
      ctx.emit('page-change', page);
    }

    onMounted(() => {
      console.log("BT Props", props);
    });
    return {
      pageChange
    };
  },
  components: { Paginator }
};
</script>