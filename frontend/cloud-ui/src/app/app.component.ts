import {Component, inject, OnInit} from '@angular/core';
import {SideBarComponent} from './components/side-bar/side-bar.component';
import {NavBarComponent} from './components/nav-bar/nav-bar.component';
import {NavigationEnd, Router, RouterOutlet, Event} from '@angular/router';
import posthog from 'posthog-js';
import {Observable, filter} from 'rxjs';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [SideBarComponent, NavBarComponent, RouterOutlet],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  title = 'Glasskube Cloud';

  private router = inject(Router);
  private navigationEnd$: Observable<NavigationEnd> = this.router.events.pipe(
    filter((event: Event) => event instanceof NavigationEnd)
  );

  public ngOnInit() {
    this.navigationEnd$.subscribe(() => posthog.capture('$pageview'));
  }
}
