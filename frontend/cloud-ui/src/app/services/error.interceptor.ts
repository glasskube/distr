import {HttpErrorResponse, HttpInterceptorFn} from '@angular/common/http';
import {tap} from 'rxjs';
import {ToastrService} from 'ngx-toastr';
import {inject} from '@angular/core';
import {ToastComponent} from '../components/toast.component';

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const toastr = inject(ToastrService);
  return next(req).pipe(
    tap({
      error: (err) => {
        if (err instanceof HttpErrorResponse) {
          toastr.error(err.statusText, 'An error occurred', {
            toastComponent: ToastComponent,
            disableTimeOut: true,
            tapToDismiss: false,
            titleClass: '',
            messageClass: '',
            toastClass: '',
            positionClass: 'toast-bottom-right',
          });
        }
      },
    })
  );
};
