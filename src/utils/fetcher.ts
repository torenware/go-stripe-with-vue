import { getTokenData } from '../logic/accounts';
import { AuthReply } from '../types/accounts';
import { JSPO } from '../types/forms';

type PagerData = {
  current_page: number;
  page_size: number;
};

type Payload = PagerData | JSPO;

export type FetchParams = {
  method: string;
  payload?: Payload | null;
};

export function NewFetchParams(payload = true) {
  // Put in defaults
  let defaultPayload: Payload;

  const params = {
    method: 'post',
  } as FetchParams;

  if (payload) {
    params.payload = {
      current_page: 1,
      page_size: 3,
    };
  }
  return params;
}

export default async function fetcher<T>(api: string, params?: FetchParams) {
  if (!params) {
    params = NewFetchParams();
  }
  const { token } = getTokenData() as AuthReply;

  // default assumes a paginated list
  let payload = params.payload;

  const requestOptions: RequestInit = {
    method: params.method,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  };

  if (['post', 'put'].includes(params.method.toLowerCase())) {
    try {
      requestOptions.body = JSON.stringify(payload);
    } catch (err) {
      console.log('could not stringify');
    }
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
