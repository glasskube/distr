import {AsyncPipe} from '@angular/common';
import {Component, inject} from '@angular/core';
import {catchError, EMPTY, map, Observable} from 'rxjs';
import {OrganizationBrandingService} from '../../services/organization-branding.service';
import {MarkdownPipe} from 'ngx-markdown';
import {HttpErrorResponse} from '@angular/common/http';
import {getFormDisplayedError} from '../../../util/errors';
import {ToastService} from '../../services/toast.service';

@Component({
  selector: 'app-home',
  imports: [AsyncPipe, MarkdownPipe],
  templateUrl: './home.component.html',
})
export class HomeComponent {
  private readonly organizationBranding = inject(OrganizationBrandingService);
  private toast = inject(ToastService);
  readonly brandingDescription$: Observable<string | undefined> = this.organizationBranding.get().pipe(
    catchError((e) => {
      const msg = getFormDisplayedError(e);
      if (msg && e instanceof HttpErrorResponse && e.status !== 404) {
        this.toast.error(msg);
      }
      return EMPTY;
    }),
    map((b) => b.description)
  );
}
