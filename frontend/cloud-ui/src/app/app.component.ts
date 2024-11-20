import {Component, inject, OnInit} from '@angular/core';
import {NavigationEnd, Router, RouterOutlet, Event} from '@angular/router';
import posthog from 'posthog-js';
import {Observable, filter} from 'rxjs';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  title = 'cloud';

  private router = inject(Router);
  private navigationEnd$: Observable<NavigationEnd> = this.router.events.pipe(
    filter((event: Event) => event instanceof NavigationEnd)
  );

  public ngOnInit() {
    this.navigationEnd$.subscribe(() => posthog.capture('$pageview'));
  }
}
