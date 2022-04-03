<template>
  <BaseTable
    :data="subscriptions"
    :headers="headers"
    :current-page="currentPage"
    :last-page="lastPage"
    :page-size="pageSize"
    :num-tabs="4"
    :total-rows="totalRows"
    @page-change="pageChange"
  >
    <template #header-row>
      <th>Order</th>
      <th>Date</th>
      <th>Item</th>
      <th>TXN ID</th>
      <th>Amount</th>
      <th>Last Four</th>
      <th>Customer</th>
      <th>Email</th>
      <th>Status</th>
    </template>
    <template #body-rows>
      <tr v-for="order in subscriptions" :key="order.id">
        <td>
          <a :href="`/admin/subscription/${order.id}`">{{ order.id }}</a>
        </td>
        <td>{{ localDate(order.created_at) }}</td>
        <td>{{ order.widget.name }}</td>
        <td>{{ order.transaction_id }}</td>
        <td>{{ formatCurrency(order.amount) }}</td>
        <td>{{ order.transaction.last_four }}</td>
        <td>{{ order.customer.last_name }}, {{ order.customer.first_name }}</td>
        <td>{{ order.customer.email }}</td>
        <td>
          <BaseBadge
            :name="badge(order.status_id).name"
            :badge-class="badge(order.status_id).class"
          ></BaseBadge>
        </td>
      </tr>
    </template>
  </BaseTable>
</template>


<script setup lang="ts">
import { onMounted, ref, Ref } from "vue";
import { format } from 'date-fns';
import { Order, PaginatedRows } from '../types/accounts';
import BaseTable from "../components/BaseTable.vue";
import BaseBadge from "../components/BaseBadge.vue";
import fetcher, { NewFetchParams } from "../utils/fetcher";

const pageSize = 3;

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

const currentPage = ref(1);
const lastPage = ref(0);
const totalRows = ref(0);


const badge = (orderStatus: number) => {
  let badgeName: string;
  let badgeClass: string;

  switch (orderStatus) {
    case 1:
      badgeName = "Subscribed";
      badgeClass = "bg-success";
      break;
    case 3:
      badgeName = "Cancelled";
      badgeClass = "bg-danger";
      break;
    default:
      badgeName = "Subscribed";
      badgeClass = "bg-success";
      break;
  }
  return {
    name: badgeName,
    class: badgeClass
  };
};

function formatCurrency(cents: number) {
  return `$${(cents / 100).toFixed(2)}`;
}

function localDate(dateStr: string) {
  const date = new Date(dateStr);
  return format(date, "yyyy-MM-dd");
}

const pageChange = (page: number) => {
  currentPage.value = page;
  updateSubs();
}

const updateSubs = async () => {
  try {
    const uri = `${window.tmpVars.api}/api/auth/list-subs`;
    const params = NewFetchParams();
    params.method = "post";
    params.payload = {
      current_page: currentPage.value,
      page_size: pageSize,
    }

    const data = await fetcher<PaginatedRows<Order>>(uri, params);
    if (!data.error) {
      const { current_page, last_page, total_rows, rows } = data as PaginatedRows<Order>;
      subscriptions.value = rows;
      currentPage.value = current_page;
      lastPage.value = last_page;
      totalRows.value = total_rows;
    }

  } catch (err) {
    console.log(err);
  }

}

onMounted(() => {
  updateSubs();
});

</script>