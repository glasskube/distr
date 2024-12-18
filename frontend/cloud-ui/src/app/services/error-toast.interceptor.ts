import {HttpErrorResponse, HttpInterceptorFn, HttpRequest} from '@angular/common/http';
import {tap} from 'rxjs';
import {inject} from '@angular/core';
import {ToastService} from './toast.service';

export const errorToastInterceptor: HttpInterceptorFn = (req, next) => {
  const toast = inject(ToastService);
  return next(req).pipe(
    tap({
      error: (err) => {
        if (err instanceof HttpErrorResponse && !ignoreError(req, err)) {
          switch (err.status) {
            case 429:
              toast.error('Rate limited! Please try again later.');
              break;
            default:
              toast.error('An error occurred');
          }
        }
      },
    })
  );
};

function ignoreError(req: HttpRequest<any>, err: HttpErrorResponse) {
  // TODO remove this; join latest deployment into the response in GET /deployments and don't do these follow up requests
  return err.status === 404 && req.url.endsWith('latest-deployment');
}
