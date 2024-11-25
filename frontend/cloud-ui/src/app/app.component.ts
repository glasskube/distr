import {Component, effect, HostBinding, inject, OnInit} from '@angular/core';
import {SideBarComponent} from './components/side-bar/side-bar.component';
import {NavBarComponent} from './components/nav-bar/nav-bar.component';
import {Event, NavigationEnd, Router, RouterOutlet} from '@angular/router';
import posthog from 'posthog-js';
import {filter, Observable} from 'rxjs';
import {ColorSchemeService} from './services/color-scheme.service';
import {FontAwesomeModule} from '@fortawesome/angular-fontawesome';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [SideBarComponent, NavBarComponent, RouterOutlet, FontAwesomeModule],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  @HostBinding('class') colorScheme: 'dark' | '' = '';
  title = 'Glasskube Cloud';

  private router = inject(Router);
  private navigationEnd$: Observable<NavigationEnd> = this.router.events.pipe(
    filter((event: Event) => event instanceof NavigationEnd)
  );

  constructor(private colorSchemeService: ColorSchemeService) {
    effect(() => {
      this.colorScheme = this.colorSchemeService.colorScheme();
    });
  }

  public ngOnInit() {
    this.navigationEnd$.subscribe(() => posthog.capture('$pageview'));
  }
}
