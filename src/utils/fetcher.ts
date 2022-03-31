import { getTokenData } from '../logic/accounts';
import { AuthReply } from '../types/accounts';

// TBD make this a setting.
const pageSize = 4;

export default async function fetcher<T>(api: string, desiredPage = 1) {
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
    return data.rows as T[];
  } catch (err) {
    console.log(err);
    return [] as T[];
  }
}
