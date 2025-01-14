import {GlobalPositionStrategy} from '@angular/cdk/overlay';
import {AsyncPipe} from '@angular/common';
import {AfterViewInit, Component, inject, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faPlus} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, first, map, Observable} from 'rxjs';
import {ApplicationsComponent} from '../../applications/applications.component';
import {DeploymentTargetsComponent} from '../../deployments/deployment-targets.component';
import {ApplicationsService} from '../../services/applications.service';
import {AuthService} from '../../services/auth.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ChartTypeComponent} from '../charts/type/chart-type.component';
import {ChartUptimeComponent} from '../charts/uptime/chart-uptime.component';
import {ChartVersionComponent} from '../charts/version/chart-version.component';
import {GlobeComponent} from '../globe/globe.component';
import {OnboardingWizardComponent} from '../onboarding-wizard/onboarding-wizard.component';
import {UsersService} from '../../services/users.service';
import {OrganizationBrandingService} from '../../services/organization-branding.service';
import {MarkdownPipe, MARKED_OPTIONS, MarkedOptions, MarkedRenderer, provideMarkdown} from 'ngx-markdown';

@Component({
  selector: 'app-home',
  imports: [AsyncPipe, MarkdownPipe],
  providers: [provideMarkdown({
    markedOptions: {
      provide: MARKED_OPTIONS,
      useFactory: markedOptionsFactory
    }
  })],
  templateUrl: './home.component.html',
})
export class HomeComponent {
  private readonly organizationBranding = inject(OrganizationBrandingService);
  readonly brandingDescription$: Observable<string | undefined> = this.organizationBranding
    .get()
    .pipe(map((b) => b.description));
}

// function that returns `MarkedOptions` with renderer override
export function markedOptionsFactory(): MarkedOptions {
  const renderer = new MarkedRenderer();

  renderer.heading = (h) => {
    console.log(h);
    return `<h${h.depth} class='text-3xl font-extrabold'>${h.text}</h${h.depth}>`;
  }

  /*renderer.code = (code) => {
    return `
        <pre>
          <code lang="${code.lang}" class="text-gray-900 dark:text-gray-200 whitespace-pre-line">
              ${code.text}
          </code>
        </pre>`
  }*/

  return {
    renderer: renderer,
  };
}
