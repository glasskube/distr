import {HttpErrorResponse, HttpInterceptorFn} from '@angular/common/http';
import {tap} from 'rxjs';
import {inject} from '@angular/core';
import {ToastService} from './toast.service';
import {displayedInToast} from '../../util/errors';
import {captureException} from '@sentry/browser';
import {captureMessage} from '@sentry/angular';

export const errorToastInterceptor: HttpInterceptorFn = (req, next) => {
  const toast = inject(ToastService);
  return next(req).pipe(
    tap({
      error: (err) => {
        const msg = getToastDisplayedError(err);
        if (msg) {
          toast.error(msg);
        }

        if (err.status && (err.status === 400 || err.status >= 500)) {
          captureException(err);
        }
      },
    })
  );
};

function getToastDisplayedError(err: any): string | undefined {
  if (displayedInToast(err) && err instanceof HttpErrorResponse) {
    switch (err.status) {
      case 429:
        const retryAfter = parseInt(err.headers.get('Retry-After') ?? '');
        if (!Number.isNaN(retryAfter)) {
          const minutes = Math.ceil(retryAfter / 60);
          return `Rate limited! Please try again in ${minutes} minute${minutes !== 1 ? 's' : ''}.`;
        }
        return 'Rate limited! Please try again later.';
      case 0:
        return 'Connection failed';
      default:
        return 'An unexpected technical error occurred';
    }
  }
  return;
}
