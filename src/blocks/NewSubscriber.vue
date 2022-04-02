<template>
  <BaseForm>
    <BaseInput id="first_name" label="First Name" />
    <BaseInput id="last_name" label="Last Name" />
    <BaseInput id="card_holder" label="Name on Card" />
    <BaseInput id="first_name" label="First Name" />
    <!-- card number field controlled by stripe js -->
    <div class="mb-3">
      <label for="card-element" class="form-label">CC Number</label>
      <div id="card-element" class="form-control"></div>
      <div id="card-errors" class="alert-danger text-center"></div>
      <div id="card-success" class="alert-success text-center"></div>
    </div>
  </BaseForm>
</template>

<script setup lang="ts">
import { ref, Ref, onMounted } from "vue";
import { useStripe } from 'vue-use-stripe';
import fetcher, { NewFetchParams } from "../utils/fetcher";
import BaseForm from "../components/BaseForm.vue";
import BaseInput from "../components/BaseInput.vue";


type StripeParams = {
  error: boolean,
  key: string
}

const sparams: Ref<StripeParams | null> = ref(null);
const stripe = ref();

onMounted(async () => {
  const params = NewFetchParams();
  params.method = "get";
  const rslt = await fetcher<StripeParams>(`${window.tmpVars.api}/api/sparams`, params);
  if (rslt.error) {
    console.log("cannot load stripe:", rslt.error);
    return;
  }
  console.log("no error", rslt);
  sparams.value = rslt as StripeParams;

  const {
    stripe: stripeObj,
    elements: [cardElement],
  } = useStripe({
    key: sparams.value.key,
    elements: [{ type: 'card', options: {} }],
  });

  stripe.value = stripeObj.value;

});

</script>