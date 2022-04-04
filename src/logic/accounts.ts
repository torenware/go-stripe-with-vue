import { JSPO } from '../types/forms';
import { AuthReply } from '../types/accounts';
import { sendFlash } from '../utils/flash';

export function handleLogin(form: HTMLFormElement, api: string, payload: JSPO) {
  const requestOptions = {
    method: 'post',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  };

  fetch(`${api}/api/authenticate`, requestOptions)
    .then((response) => response.json())
    .then((response) => {
      if (!response.error) {
        if (response.authentication_token) {
          // showLoginSuccess();
          console.log('should submit', response.authentication_token);
          loginUserToSite(response.authentication_token);
          // document.getElementById("login_form").submit();
          form.submit();
        }
      } else {
        // showLoginError(response.message);
        console.log(response.message);
      }
    });
}

export function loginUserToSite(auth_obj: AuthReply) {
  localStorage.setItem('token', auth_obj.token);
  localStorage.setItem('expiry', auth_obj.expiry);
}

export function getTokenData() {
  const token = localStorage.getItem('token');
  const expiry = localStorage.getItem('expiry');
  if (token === null || expiry === null) {
    return null;
  }
  return {
    token,
    expiry,
  };
}

export function logoutUser() {
  localStorage.removeItem('token');
  localStorage.removeItem('expiry');
  sendFlash('session expired!');
  location.href = '/logout';
}
