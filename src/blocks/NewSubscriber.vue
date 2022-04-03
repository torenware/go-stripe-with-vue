<template>
  <BaseForm>
    <BaseInput id="first_name" label="First Name" required="true" />
    <BaseInput id="last_name" label="Last Name" required="true" />
    <BaseInput id="card_holder" label="Name on Card" required="true" />
    <BaseInput id="email" input-type="email" label="Email" required="true" />

    <!-- card number field controlled by stripe js -->
    <div class="mb-3 mt-3 mx-3">
      <label for="card-element" class="form-label">CC Number</label>
      <div id="card-element" class="form-control"></div>
      <div id="card-errors" class="alert-danger text-center"></div>
      <div id="card-success" class="alert-success text-center"></div>
    </div>
  </BaseForm>
</template>

<script setup lang="ts">
import { ref, Ref, onMounted } from "vue";
import { Stripe as StripeType } from "@stripe/stripe-js/types"
import fetcher, { NewFetchParams } from "../utils/fetcher";
import BaseForm from "../components/BaseForm.vue";
import BaseInput from "../components/BaseInput.vue";


type StripeParams = {
  error: boolean,
  key: string
}

const sparams: Ref<StripeParams | null> = ref(null);
const stripe: Ref<StripeType | null> = ref(null);

const initStripeField = (stripeRef: Ref<StripeType | null>) => {
  const stripe = stripeRef.value!;
  // create stripe & elements
  const elements = stripe.elements();
  const style = {
    base: {
      fontSize: '16px',
      lineHeight: '24px'
    }
  };
  // create card entry
  let card = elements.create('card', {
    style: style,
    hidePostalCode: false,
  });
  card.mount("#card-element");

  // Stripe doesn't expose a type for this.
  type StripeError = {
    message: string;
  }

  // nor this neither ;-)
  interface StripeEvent extends Event {
    error: StripeError;
  }

  // check for input errors. card is an HTML element under
  // the hood, but the API does not expose enough methods.
  // @ts-ignore
  card.addEventListener('change', function (event: StripeEvent) {
    var displayError = document.getElementById("card-errors");
    if (!displayError) {
      console.log("expected ID card-errors");
      return;
    }
    if (event.error) {
      displayError.classList.remove('d-none');
      displayError.textContent = event.error.message;
    } else {
      displayError.classList.add('d-none');
      displayError.textContent = '';
    }
  });

}

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

  if (window.Stripe) {
    console.log('Stripe is in global space')
  } else {
    console.log("huh? no stripe?");
  }

  // Stripe is included via an injected script tag by stripe-js.js.
  // @ts-ignore
  stripe.value = Stripe(sparams.value.key);
  initStripeField(stripe)

});

</script>