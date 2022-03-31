<template>
  <BaseTable :data="subscriptions" :headers="headers">
    <template #header-row>
      <th>Order</th>
      <th>Date</th>
      <th>Item</th>
      <th>TXN ID</th>
      <th>Amount</th>
      <th>Last Four</th>
      <th>Customer</th>
      <th>Email</th>
    </template>
    <template #body-rows>
      <tr v-for="order in subscriptions" :key="order.id">
        <td>{{ order.id }}</td>
        <td>{{ localDate(order.created_at) }}</td>
        <td>{{ order.widget.name }}</td>
        <td>{{ order.transaction_id }}</td>
        <td>{{ formatCurrency(order.amount) }}</td>
        <td>{{ order.transaction.last_four }}</td>
        <td>{{ order.customer.last_name }}, {{ order.customer.first_name }}</td>
        <td>{{ order.customer.email }}</td>
      </tr>
    </template>
  </BaseTable>
</template>


<script setup lang="ts">
import { onMounted, ref, Ref } from "vue";
import { format } from 'date-fns';
import BaseTable from "../components/BaseTable.vue";
import fetcher from "../utils/fetcher";

type Widget = {
  id: number;
  name: string;
  price: number;
}

type Transaction = {
  id: number;
  currency: string;
  last_four: string;
}

type Customer = {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
}

type Order = {
  id: number;
  amount: number;
  created_at: string;
  widget: Widget;
  transaction_id: number;
  transaction: Transaction;
  customer: Customer;
}

// {
//   "id": 2,
//   "widget_id": 2,
//   "transaction_id": 2,
//   "customer_id": 2,
//   "status_id": 1,
//   "quantity": 1,
//   "amount": 2000,
//   "created_at": "2022-03-14T23:45:08Z",
//   "widget": {
//     "id": 0,
//     "name": "Bronze Plan",
//     "description": "Get three widgits per month for the price of two.",
//     "inventory_level": 0,
//     "price": 2000,
//     "is_recurring": false,
//     "plan_id": "",
//     "image": ""
//   },
//   "transaction": {
//     "id": 0,
//     "amount": 0,
//     "currency": "cad",
//     "last_four": "4242",
//     "expiry_month": 2,
//     "expiry_year": 2026,
//     "bank_return_code": "",
//     "payment_intent": "sub_1KdNZWKlT5z4v76HZ9r0pY59",
//     "payment_method": "",
//     "transaction_status_id": 0
//   },
//   "customer": {
//     "id": 0,
//     "first_name": "King",
//     "last_name": "Lir",
//     "email": "king@lir.org"
//   }
// }

const subscriptions: Ref<Order[]> = ref([]);
let headers: Ref<string[]> = ref([]);

function formatCurrency(cents: number) {
  return `$${(cents / 100).toFixed(2)}`;
}

function localDate(dateStr: string) {
  const date = new Date(dateStr);
  return format(date, "yyyy-MM-dd");
}

onMounted(async () => {
  try {
    const data = await fetcher<Order>("/api/auth/list-subs");
    subscriptions.value = data;
    console.log(subscriptions);
    headers.value = ["ID", "LAST", "EMAIL"];

  } catch (err) {
    console.log(err);
  }
});

</script>