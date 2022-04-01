import { getTokenData } from '../logic/accounts';
import { AuthReply } from '../types/accounts';

export default async function fetcher<T>(
  api: string,
  desiredPage = 1,
  pageSize = 3
) {
  let rows = [];
  const { token } = getTokenData() as AuthReply;
  const payload = {
    page_size: pageSize,
    current_page: desiredPage,
  };

  const requestOptions = {
    method: 'post',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  };
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
