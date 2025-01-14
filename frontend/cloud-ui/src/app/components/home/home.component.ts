import {AsyncPipe} from '@angular/common';
import {Component, inject} from '@angular/core';
import {map, Observable} from 'rxjs';
import {OrganizationBrandingService} from '../../services/organization-branding.service';
import {MarkdownPipe, MARKED_OPTIONS, provideMarkdown} from 'ngx-markdown';
import {markedOptionsFactory} from '../../services/markdown-options.factory';

@Component({
  selector: 'app-home',
  imports: [AsyncPipe, MarkdownPipe],
  providers: [
    provideMarkdown({
      markedOptions: {
        provide: MARKED_OPTIONS,
        useFactory: markedOptionsFactory,
      },
    }),
  ],
  templateUrl: './home.component.html',
})
export class HomeComponent {
  private readonly organizationBranding = inject(OrganizationBrandingService);
  readonly brandingDescription$: Observable<string | undefined> = this.organizationBranding
    .get()
    .pipe(map((b) => b.description));
}
