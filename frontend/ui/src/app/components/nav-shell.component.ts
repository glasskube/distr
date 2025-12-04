import {Component, computed, inject, signal} from '@angular/core';
import {toSignal} from '@angular/core/rxjs-interop';
import {RouterOutlet} from '@angular/router';
import dayjs from 'dayjs';
import {map} from 'rxjs';
import {AuthService} from '../services/auth.service';
import {OrganizationService} from '../services/organization.service';
import {NavBarComponent} from './nav-bar/nav-bar.component';
import {SideBarComponent} from './side-bar/side-bar.component';

@Component({
  selector: 'app-nav-shell',
  template: `
    <app-nav-bar [isSubscriptionBannerVisible]="isSubscriptionBannerVisible()"></app-nav-bar>
    @if (!isSubscriptionExpired()) {
      <app-side-bar [isSubscriptionBannerVisible]="isSubscriptionBannerVisible()"></app-side-bar>
    }
    <div [class.sm:ml-64]="!isSubscriptionExpired()">
      <router-outlet></router-outlet>
    </div>
  `,
  imports: [NavBarComponent, SideBarComponent, RouterOutlet],
})
export class NavShellComponent {
  private readonly organizationService = inject(OrganizationService);
  private readonly auth = inject(AuthService);

  private readonly isVendor = signal(this.auth.isVendor());
  protected readonly organization$ = this.organizationService.get();

  protected readonly isSubscriptionExpired = toSignal(
    this.organization$.pipe(map((org) => dayjs(org.subscriptionEndsAt).isBefore())),
    {initialValue: false}
  );

  private readonly isSubscriptionTrial = toSignal(
    this.organization$.pipe(map((org) => org.subscriptionType === 'trial')),
    {initialValue: false}
  );

  protected readonly isSubscriptionBannerVisible = computed(() => {
    if (!this.isVendor()) {
      return false;
    }
    return this.isSubscriptionExpired() || this.isSubscriptionTrial();
  });
}
