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
  authenticate?: boolean;
};

export type FetchError = {
  error: string;
  status?: number;
};

export function NewFetchParams(payload = true) {
  const params = {
    method: 'post',
    authenticate: true,
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

  // default assumes a paginated list
  let payload = params.payload;
  // const headers = new Headers();
  // headers.set('Accept', 'application/json');
  // headers.set('Content-Type', 'application/json');
  const headers: JSPO = {
    Accept: 'application/json',
    'Content-Type': 'application/json',
  };
  if (params.authenticate) {
    const tokenData = getTokenData();
    if (tokenData) {
      const { token } = tokenData as AuthReply;
      headers['Authorization'] = `Bearer ${token}`;
    } else {
      console.log('token data was unavailable');
      return {
        error: 'session is expired or not set',
        status: 401,
      };
    }
  }

  // const processedHeaders: JSPO = {};
  // headers.forEach((val, key) => {
  //   processedHeaders[key] = val;
  // });

  const requestOptions: RequestInit = {
    method: params.method,
    headers: headers as HeadersInit,
  };

  console.log(requestOptions);

  if (['post', 'put'].includes(params.method.toLowerCase())) {
    try {
      requestOptions.body = JSON.stringify(payload);
    } catch (err) {
      console.log('could not stringify');
    }
  }

  // placating typescript here:
  let rslt: Response = new Response();
  try {
    rslt = await fetch(api, requestOptions);
    const data = await rslt.json();
    return data as T;
  } catch (err) {
    console.log(err);
    return {
      error: err,
      //@ts-ignore
      status: rslt.status,
    } as FetchError;
  }
}
