import {HttpErrorResponse, HttpInterceptorFn, HttpRequest} from '@angular/common/http';
import {tap} from 'rxjs';
import {inject} from '@angular/core';
import {ToastService} from './toast.service';
import {displayedInToast} from '../../util/errors';

export const errorToastInterceptor: HttpInterceptorFn = (req, next) => {
  const toast = inject(ToastService);
  return next(req).pipe(
    tap({
      error: (err) => {
        const msg = getToastDisplayedError(err);
        if(msg) {
          toast.error(msg);
        }
      },
    })
  );
};

function getToastDisplayedError(err: any): string | undefined {
  if (displayedInToast(err) && err instanceof HttpErrorResponse) {
    switch(err.status) {
      case 429: return 'Rate limited! Please try again later.';
      default: return 'An unexpected technical error occurred';
    }
  }
  return;
}
