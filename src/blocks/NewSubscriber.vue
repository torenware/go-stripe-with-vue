<template>
  <BaseForm reset-text="Reset" :process="processCard" @reset="handleReset">
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
import type { PaymentMethodResult, Stripe as StripeType, StripeCardElement } from "@stripe/stripe-js/types";
import { loadStripe } from "@stripe/stripe-js";
import fetcher, { NewFetchParams, FetchError } from "../utils/fetcher";
import { sendFlash } from "../utils/flash";
import { ProcessSubmitFunc, JSPO } from "../types/forms";
import BaseForm from "../components/BaseForm.vue";
import BaseInput from "../components/BaseInput.vue";

type Widget = {
  id: number,
  name: string,
  price: number,
  plan_id: string,
  is_recurring: boolean,
  description: string,
}

type StripeParams = {
  error: boolean,
  key: string,
  widget: Widget,
}

const sparams: Ref<StripeParams | null> = ref(null);
const stripe: Ref<StripeType | null> = ref(null);
const cardField: Ref<StripeCardElement | null> = ref(null);

const processCard: ProcessSubmitFunc = async (data, form) => {
  let intent: PaymentMethodResult | undefined;
  try {
    intent = await stripe.value?.createPaymentMethod({
      type: "card",
      card: cardField.value!,
      billing_details: {
        email: data["email"] as string
      }
    });
    await stripePaymentMethodHandler(intent!, data)
  } catch (err) {
    sendFlash("Sorry. Could not create your subscription");
    console.log("failed:", err);
  }
}

async function stripePaymentMethodHandler(rslt: PaymentMethodResult, data: JSPO) {
  type SubscriptionReply = {
    ok: boolean,
    message: string,
  }

  if (rslt.error) {
    sendFlash("could not complete subscription")
  } else {
    const params = sparams.value! as StripeParams;
    // create customer and subscribe.
    let payload = {
      product_id: params.widget.id,
      plan: params.widget.plan_id,
      payment_method: rslt.paymentMethod.id,
      email: data.email,
      last_four: rslt.paymentMethod.card?.last4 as string,
      card_brand: rslt.paymentMethod.card?.brand as string,
      exp_month: rslt.paymentMethod.card?.exp_month as number,
      exp_year: rslt.paymentMethod.card?.exp_year as number,
      first_name: data.first_name as string,
      last_name: data.last_name as string,
      amount: params.widget.price as number,
      currency: "cad", // Put into widget table
    };

    const uri = `${window.tmpVars.api}/api/create-customer-and-subscribe-to-plan`;
    const fetchParams = NewFetchParams();
    fetchParams.payload = payload
    const subRslt = await fetcher<SubscriptionReply>(uri, fetchParams)
    // todo: either use an error or an ok but not both on
    // the server.
    if (!(subRslt as SubscriptionReply).ok) {
      let msg = (subRslt as SubscriptionReply).message;
      if (!msg) {
        const fetchErr = subRslt as FetchError;
        msg = fetchErr.error;
      }
      throw new Error(msg);
    }

    // Stuff our data into session_storage
    sessionStorage.setItem("first_name", payload.first_name)
    sessionStorage.setItem("last_name", payload.last_name)
    sessionStorage.setItem("amount", (payload.amount / 100).toFixed(2))
    sessionStorage.setItem("last_four", payload.last_four)
    sessionStorage.setItem("card_brand", payload.card_brand)
    sessionStorage.setItem("item", params.widget.name)
    sessionStorage.setItem("description", params.widget.description)

    location.href = "/receipt/bronze";

  }
}


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
  cardField.value = card;

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

const handleReset = () => {
  if (cardField.value) {
    cardField.value.clear();
  }
}

onMounted(async () => {
  const params = NewFetchParams();
  params.method = "get";
  const rslt = await fetcher<StripeParams>(`${window.tmpVars.api}/api/sparams/2`, params);
  if (rslt.error) {
    console.log("cannot load stripe:", rslt.error);
    return;
  }
  sparams.value = rslt as StripeParams;

  if (window.Stripe) {
    console.log('Stripe is in global space');
    stripe.value = window.Stripe(sparams.value.key);

  } else {
    try {
      stripe.value = await loadStripe(sparams.value.key);
      console.log("late load for Stripe");
    } catch (err) {
      console.log("huh? no stripe?", err);
    }
  }

  // Stripe is included via an injected script tag by stripe-js.js.
  // @ts-ignore
  initStripeField(stripe)

});

</script>