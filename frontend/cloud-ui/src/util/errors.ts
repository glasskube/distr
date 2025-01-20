import {HttpErrorResponse} from '@angular/common/http';

export function displayedInToast(err: any): boolean {
  if (err instanceof HttpErrorResponse) {
    return !err.status || err.status === 429 || err.status >= 500;
  }
  return false;
}

export function getFormDisplayedError(err: any): string | undefined {
  if (!displayedInToast(err)) {
    if (err instanceof HttpErrorResponse && typeof err.error === 'string') {
      return err.error;
    } else {
      return 'Something went wrong';
    }
  }
  return;
}
