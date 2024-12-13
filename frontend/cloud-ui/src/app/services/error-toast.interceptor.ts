import {HttpErrorResponse, HttpInterceptorFn} from '@angular/common/http';
import {tap} from 'rxjs';
import {inject} from '@angular/core';
import {ToastService} from './toast.service';

export const errorToastInterceptor: HttpInterceptorFn = (req, next) => {
  const toast = inject(ToastService);
  return next(req).pipe(
    tap({
      error: (err) => {
        if (err instanceof HttpErrorResponse) {
          toast.error('An error occurred');
        }
      },
    })
  );
};
