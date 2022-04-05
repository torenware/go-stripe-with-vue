const tmpl = `
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
</template>`;

export default function componentDef<T>() {
  return {
    template: tmpl,
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
      return {};
    },
  };
}
