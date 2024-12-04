import {Component, effect, HostBinding, inject, OnInit} from '@angular/core';
import {SideBarComponent} from './components/side-bar/side-bar.component';
import {NavBarComponent} from './components/nav-bar/nav-bar.component';
import {Event, NavigationEnd, Router, RouterOutlet} from '@angular/router';
import posthog from 'posthog-js';
import {filter, interval, Observable} from 'rxjs';
import {ColorSchemeService} from './services/color-scheme.service';
import {FontAwesomeModule} from '@fortawesome/angular-fontawesome';
import {initFlowbite} from 'flowbite';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, FontAwesomeModule],
  template: `<router-outlet></router-outlet>`,
})
export class AppComponent implements OnInit {
  private readonly router = inject(Router);
  private readonly navigationEnd$: Observable<NavigationEnd> = this.router.events.pipe(
    filter((event: Event) => event instanceof NavigationEnd)
  );

  constructor(private readonly colorSchemeService: ColorSchemeService) {
    effect(() => {
      document.body.classList.toggle('dark', this.colorSchemeService.colorScheme() === 'dark');
    });
  }

  public ngOnInit() {
    this.navigationEnd$.subscribe(() => posthog.capture('$pageview'));
  }
}
