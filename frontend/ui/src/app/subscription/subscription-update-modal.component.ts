import {CurrencyPipe} from '@angular/common';
import {Component, inject, input, output} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faCheck, faXmark} from '@fortawesome/free-solid-svg-icons';
import {getFormDisplayedError} from '../../util/errors';
import {modalFlyInOut} from '../animations/modal';
import {SubscriptionService} from '../services/subscription.service';
import {ToastService} from '../services/toast.service';
import {SubscriptionInfo, SubscriptionPeriode} from '../types/subscription';

export interface PendingSubscriptionUpdate {
  userAccountQuantity: number;
  customerOrganizationQuantity: number;
  newPrice: number;
  oldPrice: number;
  subscriptionPeriode: SubscriptionPeriode;
}

@Component({
  selector: 'app-subscription-update-modal',
  templateUrl: './subscription-update-modal.component.html',
  imports: [FaIconComponent, CurrencyPipe],
  animations: [modalFlyInOut],
})
export class SubscriptionUpdateModalComponent {
  protected readonly xmarkIcon = faXmark;
  protected readonly checkIcon = faCheck;

  private readonly subscriptionService = inject(SubscriptionService);
  private readonly toast = inject(ToastService);

  pendingUpdate = input.required<PendingSubscriptionUpdate>();
  closed = output<void>();
  confirmed = output<SubscriptionInfo>();

  async confirmUpdate() {
    const pending = this.pendingUpdate();
    if (!pending) {
      return;
    }

    try {
      const body = {
        subscriptionUserAccountQuantity: pending.userAccountQuantity,
        subscriptionCustomerOrganizationQuantity: pending.customerOrganizationQuantity,
      };

      const updatedInfo = await this.subscriptionService.updateSubscription(body);
      this.toast.success('Subscription updated successfully');
      this.confirmed.emit(updatedInfo);
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
  }

  close() {
    this.closed.emit();
  }
}
