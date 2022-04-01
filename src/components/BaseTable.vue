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
      :num-tabs="numTabs"
      :page-size="pageSize"
    ></Paginator>
  </section>
</template>

<script>
import Paginator from "../components/Paginator.vue";

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
    numTabs: {
      type: Number,
      required: false,
      default: 4,
    },
    pageSize: {
      type: Number,
      required: false,
      default: 20,
    }

  },
  setup(props, ctx) {

    function pageChange(page) {
      ctx.emit('page-change', page);
    }

    return {
      pageChange
    };
  },
  components: { Paginator }
};
</script>