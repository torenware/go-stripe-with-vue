import { FlashData } from '../types/forms';

export const sendFlash = (msg: string, alertType: string = 'alert-danger') => {
  const payload: CustomEventInit<FlashData> = {
    detail: {
      msg,
      alertType,
    },
  };
  const evt = new CustomEvent<FlashData>('flashMsg', payload);
  const flashPanel = document.querySelector('#flashPanel');
  if (flashPanel) {
    flashPanel.dispatchEvent(evt);
  } else {
    console.log('cannot find the #flashPanel');
  }
};
