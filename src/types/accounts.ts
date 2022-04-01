import intervalToDuration from 'date-fns/intervalToDuration';

export type AuthReply = {
  token: string;
  expiry: string;
};

export type PaginatedRows<T> = {
  error: boolean;
  current_page: number;
  last_page: number;
  total_rows: number;
  rows: T[];
};

export type SingleItem<T> = {
  error: string;
  message: string;
  item: T;
};

export type ServerError = {
  error: boolean;
  message: string;
};

export type Widget = {
  id: number;
  name: string;
  price: number;
};

export type Transaction = {
  id: number;
  currency: string;
  last_four: string;
};

export type Customer = {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
};

export type Order = {
  id: number;
  amount: number;
  created_at: string;
  widget: Widget;
  transaction_id: number;
  transaction: Transaction;
  customer: Customer;
  status_id: number;
};
