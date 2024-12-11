import {HttpErrorResponse, HttpInterceptorFn, HttpRequest} from '@angular/common/http';
import {tap} from 'rxjs';
import {IndividualConfig, ToastrService} from 'ngx-toastr';
import {inject} from '@angular/core';
import {ToastComponent} from '../components/toast.component';

export const GLOBAL_ERROR_TOAST_CONFIG: Partial<IndividualConfig> = {
  toastComponent: ToastComponent,
  disableTimeOut: true,
  tapToDismiss: false,
  titleClass: '',
  messageClass: '',
  toastClass: '',
  positionClass: 'toast-bottom-right',
};

export const errorToastInterceptor: HttpInterceptorFn = (req, next) => {
  const toastr = inject(ToastrService);
  return next(req).pipe(
    tap({
      error: (err) => {
        if (err instanceof HttpErrorResponse) {
          if (!ignoreGlobalError(req, err)) {
            toastr.show(err.statusText, 'An error occurred', GLOBAL_ERROR_TOAST_CONFIG);
          }
        }
      },
    })
  );
};

function ignoreGlobalError(req: HttpRequest<any>, err: HttpErrorResponse): boolean {
  if (err.status >= 400 && err.status < 500) {
    return true;
  }
  return false;
}
