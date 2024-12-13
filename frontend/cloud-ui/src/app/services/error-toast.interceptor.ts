import {HttpErrorResponse, HttpInterceptorFn, HttpRequest} from '@angular/common/http';
import {tap} from 'rxjs';
import {IndividualConfig, ToastrService} from 'ngx-toastr';
import {inject} from '@angular/core';
import {ToastComponent} from '../components/toast.component';
import {ToastService} from './toast.service';

export const errorToastInterceptor: HttpInterceptorFn = (req, next) => {
  const toast = inject(ToastService);
  return next(req).pipe(
    tap({
      error: (err) => {
        if (err instanceof HttpErrorResponse) {
          if (!ignoreGlobalError(req, err)) {
            toast.error('An internal server error occurred');
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
