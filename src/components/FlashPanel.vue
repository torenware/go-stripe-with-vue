<template>
  <div class="row">
    <div class="col" style="width: 100%">
      <div
        :id="id"
        class="alert alert-danger text-center mx-auto mt-3"
        :class="flashClass"
        ref="flashPanel"
      >{{ flashMsg }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, Ref, onMounted } from 'vue';
import Cookies from 'js-cookie';
import { FlashData } from "../types/forms";

const props = withDefaults(defineProps<{
  id?: string
}>(), {
  id: "flashPanel"
});

const flashPanel: Ref<Element | null> = ref(null);
const flashMsg = ref("");
const flashClass = ref("d-none alert-danger");

const showFlash = (msg: string, alertType: string = "") => {
  flashMsg.value = msg;
  flashClass.value = alertType === "" ? "alert-danger" : alertType;

  setTimeout(() => {
    hideFlash();
  }, 10 * 1000);
};

const hideFlash = () => {
  flashMsg.value = "";
  flashClass.value = "d-none alert-danger";
}

onMounted(() => {
  document.addEventListener("DOMContentLoaded", evt => {
    const flashVal = Cookies.get("flash");
    const flashType = Cookies.get("flash-type")

    let typeClass = "alert-danger";
    if (flashType) {
      typeClass = `alert-${decodeURIComponent(atob(flashType))}`;
    }

    if (flashVal) {
      const decoded = decodeURIComponent(atob(flashVal));
      showFlash(decoded, typeClass);
      Cookies.remove("flash");
    }
  });

  // Use a custom event to start a flash from JS
  flashPanel.value?.addEventListener("flashMsg", evt => {
    const customEvt = evt as CustomEvent<FlashData>;
    const { msg, alertType } = customEvt.detail;
    showFlash(msg, alertType);
  });

});

</script>

<style>
div.alert {
  width: 50%;
}
</style>