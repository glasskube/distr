import {Component, effect, inject} from '@angular/core';
import {CommonModule} from '@angular/common';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faCheckCircle} from '@fortawesome/free-solid-svg-icons';
import {RouterLink} from '@angular/router';
import {OrganizationService} from '../services/organization.service';
import {toSignal} from '@angular/core/rxjs-interop';
import {map} from 'rxjs';

@Component({
  selector: 'app-subscription-callback',
  templateUrl: './subscription-callback.component.html',
  imports: [CommonModule, FaIconComponent, RouterLink],
})
export class SubscriptionCallbackComponent {
  private readonly organizationService = inject(OrganizationService);

  protected readonly faCheckCircle = faCheckCircle;

  protected readonly isTrial = toSignal(
    this.organizationService.get().pipe(map((org) => org.subscriptionType === 'trial'))
  );

  constructor() {
    effect(() => {
      if (this.isTrial()) {
        setTimeout(() => location.reload(), 5000);
      }
    });
  }
}
