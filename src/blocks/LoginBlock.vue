<template>
  <BaseForm :process="processLogin" action="/process-login" method="post" resetText="Reset Form">
    <template #default="fromForm">
      <BaseInput label="Email" id="email" name="email" inputType="email" required="true" />
      <BaseInput
        label="Password"
        id="password"
        name="password"
        inputType="password"
        required="true"
      />
    </template>
  </BaseForm>
</template>

<script setup lang="ts">
import { JSPO, ProcessSubmitFunc } from "../types/forms";
import BaseForm from "../components/BaseForm.vue";
import BaseInput from "../components/BaseInput.vue";
import { handleLogin } from "../logic/accounts";


const processLogin: ProcessSubmitFunc = (data: JSPO, form: HTMLFormElement | null) => {
  console.log("submitting data", data);
  if (!form) {
    console.log("form isn't ready");
    return;
  }
  type importedData = {
    api: string;
    uid: number;
  }

  // @ts-ignore
  const { api } = (tmpVars as importedData);
  if (!api) {
    console.log("no api data");
  }
  handleLogin(form, api, data);
}

</script>

<style scoped>
</style>