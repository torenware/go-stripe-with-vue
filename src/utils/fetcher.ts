import { getTokenData } from '../logic/accounts';
import { AuthReply } from '../types/accounts';

export type FetchParams = {
  method: string;
  page: number;
  pageSize: number;
};

export function NewFetchParams() {
  // Put in defaults
  return {
    method: 'post',
    page: 1,
    pageSize: 3,
  } as FetchParams;
}

export default async function fetcher<T>(
  api: string,
  params?: FetchParams
  // desiredPage = 1,
  // pageSize = 3
) {
  if (!params) {
    params = NewFetchParams();
  }
  const { token } = getTokenData() as AuthReply;

  const payload = {
    page_size: params.pageSize,
    current_page: params.page,
  };

  const requestOptions: RequestInit = {
    method: params.method,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  };

  if (params.method.toLowerCase() in ['post', 'put']) {
    requestOptions.body = JSON.stringify(payload);
  }

  try {
    const rslt = await fetch(api, requestOptions);
    const data = await rslt.json();
    return data as T;
  } catch (err) {
    console.log(err);
    return {
      error: err,
    };
  }
}
