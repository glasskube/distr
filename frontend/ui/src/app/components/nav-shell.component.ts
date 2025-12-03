import {Component, inject} from '@angular/core';
import {RouterOutlet} from '@angular/router';
import {NavBarComponent} from './nav-bar/nav-bar.component';
import {SideBarComponent} from './side-bar/side-bar.component';
import {toSignal} from '@angular/core/rxjs-interop';
import dayjs from 'dayjs';
import {map} from 'rxjs';
import {OrganizationService} from '../services/organization.service';

@Component({
  selector: 'app-nav-shell',
  template: `
    <app-nav-bar></app-nav-bar>
    @if (!isSubscriptionExpired()) {
      <app-side-bar></app-side-bar>
    }
    <div [class.sm:ml-64]="!isSubscriptionExpired()">
      <router-outlet></router-outlet>
    </div>
  `,
  imports: [NavBarComponent, SideBarComponent, RouterOutlet],
})
export class NavShellComponent {
  private readonly organizationService = inject(OrganizationService);

  protected readonly organization$ = this.organizationService.get();

  protected readonly isSubscriptionExpired = toSignal(
    this.organization$.pipe(map((org) => dayjs(org.subscriptionEndsAt).isBefore()))
  );
}
