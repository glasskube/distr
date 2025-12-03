import {CurrencyPipe} from '@angular/common';
import {Component, EventEmitter, inject, Input, Output} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faCheck, faXmark} from '@fortawesome/free-solid-svg-icons';
import {getFormDisplayedError} from '../../util/errors';
import {modalFlyInOut} from '../animations/modal';
import {SubscriptionService} from '../services/subscription.service';
import {ToastService} from '../services/toast.service';
import {SubscriptionInfo} from '../types/subscription';

export interface PendingSubscriptionUpdate {
  userAccountQuantity: number;
  customerOrganizationQuantity: number;
  newPrice: number;
  oldPrice: number;
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

  @Input() pendingUpdate?: PendingSubscriptionUpdate;
  @Output('closed') readonly closed = new EventEmitter<void>();
  @Output('confirmed') readonly confirmed = new EventEmitter<SubscriptionInfo>();

  async confirmUpdate() {
    if (!this.pendingUpdate) {
      return;
    }

    try {
      const body = {
        subscriptionUserAccountQuantity: this.pendingUpdate.userAccountQuantity,
        subscriptionCustomerOrganizationQuantity: this.pendingUpdate.customerOrganizationQuantity,
      };

      const updatedInfo = await this.subscriptionService.updateSubscription(body);
      this.toast.success('Subscription updated successfully');
      this.confirmed.emit(updatedInfo);
      this.close();
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
