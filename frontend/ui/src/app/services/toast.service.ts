import {inject, Injectable} from '@angular/core';
import {IndividualConfig, ToastrService} from 'ngx-toastr';
import {ToastComponent} from '../components/toast.component';

const toastBaseConfig: Partial<IndividualConfig> = {
  toastComponent: ToastComponent,
  disableTimeOut: true,
  tapToDismiss: false,
  titleClass: '',
  messageClass: '',
  toastClass: '',
  positionClass: 'toast-bottom-right',
};

export type ToastType = 'success' | 'error' | 'info';

@Injectable({providedIn: 'root'})
export class ToastService {
  private readonly toastr = inject(ToastrService);

  public success(message: string) {
    this.toastr.show<ToastType>('', message, {
      ...toastBaseConfig,
      payload: 'success',
      disableTimeOut: false,
    });
  }

  public error(message: string) {
    this.toastr.show('', message, {
      ...toastBaseConfig,
      payload: 'error',
    });
  }

  public info(message: string) {
    return this.toastr.show<ToastType>('', message, {
      ...toastBaseConfig,
      payload: 'info',
      disableTimeOut: true,
    });
  }
}
