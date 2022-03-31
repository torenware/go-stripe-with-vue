<template>
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
</template>

<script lang="ts">
import { defineComponent } from "vue";

type TabDataItem = {
  id: string;
  first: string;
  last: string;
}


// @see https://gist.github.com/GavinRay97/29f9af2f84665e5656e8298a568f1175
function componentDef<T>() {
  return {
    props: {
      data: {
        type: Object as () => T[],
        required: true,
      },
      headers: {
        type: Object as () => string[],
        required: true,
      },
    },
    setup(/** props, context **/) {
      return {

      }
    }
  }
}

export default defineComponent(componentDef<TabDataItem>());


</script>