import {Component, effect, inject, OnInit} from '@angular/core';
import {Event, NavigationEnd, Router, RouterOutlet} from '@angular/router';
import posthog from 'posthog-js';
import {filter, Observable} from 'rxjs';
import {ColorSchemeService} from './services/color-scheme.service';
import {FontAwesomeModule} from '@fortawesome/angular-fontawesome';
import * as Sentry from '@sentry/angular';
import {AuthService} from './services/auth.service';
import {OverlayService} from './services/overlay.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, FontAwesomeModule],
  providers: [OverlayService],
  template: `<router-outlet></router-outlet>`,
})
export class AppComponent implements OnInit {
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);
  private readonly navigationEnd$: Observable<NavigationEnd> = this.router.events.pipe(
    filter((event: Event) => event instanceof NavigationEnd)
  );

  constructor(private readonly colorSchemeService: ColorSchemeService) {
    effect(() => {
      document.body.classList.toggle('dark', this.colorSchemeService.colorScheme() === 'dark');
    });
  }

  public ngOnInit() {
    this.navigationEnd$.subscribe(() => {
      const email = this.auth.getClaims()?.email;
      Sentry.setUser({email});
      posthog.setPersonProperties({email});
      posthog.capture('$pageview');
    });
  }
}
