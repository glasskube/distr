import {GlobalPositionStrategy, OverlayModule} from '@angular/cdk/overlay';
import {CommonModule} from '@angular/common';
import {Component, computed, inject, OnInit, signal, TemplateRef, ViewChild} from '@angular/core';
import {toSignal} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faCreditCard, faShoppingCart} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom} from 'rxjs';
import {getFormDisplayedError} from '../../util/errors';
import {never} from '../../util/exhaust';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {SubscriptionService} from '../services/subscription.service';
import {ToastService} from '../services/toast.service';
import {SubscriptionInfo, SubscriptionType} from '../types/subscription';
import {PendingSubscriptionUpdate, SubscriptionUpdateModalComponent} from './subscription-update-modal.component';

@Component({
  selector: 'app-subscription',
  templateUrl: './subscription.component.html',
  imports: [FaIconComponent, ReactiveFormsModule, CommonModule, OverlayModule, SubscriptionUpdateModalComponent],
})
export class SubscriptionComponent implements OnInit {
  protected readonly faShoppingCart = faShoppingCart;
  protected readonly faCreditCard = faCreditCard;

  private readonly subscriptionService = inject(SubscriptionService);
  private readonly toast = inject(ToastService);
  private readonly overlay = inject(OverlayService);

  protected subscriptionInfo = signal<SubscriptionInfo | undefined>(undefined);
  protected pendingUpdate = signal<PendingSubscriptionUpdate | undefined>(undefined);

  private modal?: DialogRef;

  @ViewChild('updateModal') protected readonly updateModal!: TemplateRef<unknown>;

  protected readonly form = new FormGroup({
    subscriptionType: new FormControl<SubscriptionType>('pro', [Validators.required]),
    billingMode: new FormControl<'monthly' | 'yearly'>('monthly', [Validators.required]),
    userAccountQuantity: new FormControl<number>(1, [Validators.required, Validators.min(1)]),
    customerOrganizationQuantity: new FormControl<number>(1, [Validators.required, Validators.min(0)]),
  });

  protected readonly formValues = toSignal(this.form.valueChanges, {initialValue: this.form.value});

  protected readonly hasQuantitiesChanged = computed(() => {
    const info = this.subscriptionInfo();
    const values = this.formValues();

    if (!info) {
      return false;
    }

    return (
      values.userAccountQuantity !== info.subscriptionUserAccountQuantity ||
      values.customerOrganizationQuantity !== info.subscriptionCustomerOrganizationQuantity
    );
  });

  async ngOnInit() {
    try {
      const info = await firstValueFrom(this.subscriptionService.get());
      this.subscriptionInfo.set(info);

      // Pre-fill form with current subscription values or defaults
      this.form.patchValue({
        subscriptionType: info.subscriptionType === 'trial' ? 'pro' : info.subscriptionType,
        userAccountQuantity: info.subscriptionUserAccountQuantity ?? info.currentUserAccountCount,
        customerOrganizationQuantity:
          info.subscriptionCustomerOrganizationQuantity ?? info.currentCustomerOrganizationCount,
      });
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
  }

  getPreviewPrice(): number {
    const subscriptionType = this.form.value.subscriptionType;
    const billingMode = this.form.value.billingMode;
    const userQty = this.form.value.userAccountQuantity ?? 0;
    const customerQty = this.form.value.customerOrganizationQuantity ?? 0;

    let userPrice = 0;
    let customerPrice = 0;

    if (subscriptionType === 'starter') {
      userPrice = billingMode === 'monthly' ? 19 : 192;
      customerPrice = billingMode === 'monthly' ? 29 : 288;
    } else if (subscriptionType === 'pro') {
      userPrice = billingMode === 'monthly' ? 29 : 288;
      customerPrice = billingMode === 'monthly' ? 69 : 672;
    }

    return userPrice * userQty + customerPrice * customerQty;
  }

  async checkout() {
    this.form.markAllAsTouched();
    if (this.form.valid) {
      try {
        const body = {
          subscriptionType: this.form.value.subscriptionType!,
          billingMode: this.form.value.billingMode!,
          subscriptionUserAccountQuantity: this.form.value.userAccountQuantity!,
          subscriptionCustomerOrganizationQuantity: this.form.value.customerOrganizationQuantity!,
        };

        // Call the checkout endpoint which will redirect to Stripe
        await this.subscriptionService.checkout(body);
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      }
    }
  }

  async updateQuantities() {
    this.form.markAllAsTouched();
    if (this.form.valid) {
      const info = this.subscriptionInfo();
      if (!info) {
        return;
      }

      // Calculate current and new prices
      const oldPrice = this.calculateCurrentPrice();
      const newPrice = this.getPreviewPrice();

      // Set pending update and show confirmation modal
      this.pendingUpdate.set({
        userAccountQuantity: this.form.value.userAccountQuantity!,
        customerOrganizationQuantity: this.form.value.customerOrganizationQuantity!,
        newPrice,
        oldPrice,
      });

      this.hideModal();
      this.modal = this.overlay.showModal(this.updateModal, {
        hasBackdrop: true,
        backdropStyleOnly: true,
        positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
      });
    }
  }

  onModalConfirmed(updatedInfo: SubscriptionInfo) {
    this.subscriptionInfo.set(updatedInfo);
    this.hideModal();
  }

  hideModal() {
    this.modal?.close();
  }

  private calculateCurrentPrice(): number {
    const info = this.subscriptionInfo();
    if (!info || !info.subscriptionUserAccountQuantity || !info.subscriptionCustomerOrganizationQuantity) {
      return 0;
    }

    const subscriptionType = info.subscriptionType;
    const userQty = info.subscriptionUserAccountQuantity;
    const customerQty = info.subscriptionCustomerOrganizationQuantity;

    let userPrice = 0;
    let customerPrice = 0;

    // Assume monthly billing for current price (adjust if you have billing mode info)
    if (subscriptionType === 'starter') {
      userPrice = 19;
      customerPrice = 29;
    } else if (subscriptionType === 'pro') {
      userPrice = 29;
      customerPrice = 69;
    }

    return userPrice * userQty + customerPrice * customerQty;
  }

  getPlanLimits(plan: SubscriptionType): {customers: string; users: string; deployments: string} {
    const limits = this.getPlanLimitsObject(plan);
    if (!limits) {
      return {customers: '', users: '', deployments: ''};
    }

    return {
      customers:
        limits.maxCustomerOrganizations === -1
          ? 'Unlimited customer organizations'
          : `Up to ${limits.maxCustomerOrganizations} customer organization${limits.maxCustomerOrganizations > 1 ? 's' : ''}`,
      users:
        limits.maxUsersPerCustomerOrganization === -1
          ? 'Unlimited users per customer organization'
          : `Up to ${limits.maxUsersPerCustomerOrganization} user account${limits.maxUsersPerCustomerOrganization > 1 ? 's' : ''} per customer organization`,
      deployments:
        limits.maxDeploymentsPerCustomerOrganization === -1
          ? 'Unlimited deployments per customer'
          : `${limits.maxDeploymentsPerCustomerOrganization} active deployment${limits.maxDeploymentsPerCustomerOrganization > 1 ? 's' : ''} per customer`,
    };
  }

  private getPlanLimitsObject(subscriptionType: SubscriptionType) {
    const info = this.subscriptionInfo();
    if (!info) {
      return null;
    }

    switch (subscriptionType) {
      case 'trial':
        return info.trialLimits;
      case 'starter':
        return info.starterLimits;
      case 'pro':
        return info.proLimits;
      case 'enterprise':
        return info.enterpriseLimits;
      default:
        return never(subscriptionType);
    }
  }

  getPlanLimit(
    subscriptionType: SubscriptionType,
    metric: 'customerOrganizations' | 'usersPerCustomer' | 'deploymentsPerCustomer'
  ): string | number {
    const limits = this.getPlanLimitsObject(subscriptionType);
    if (!limits) {
      return '';
    }

    switch (metric) {
      case 'customerOrganizations':
        return limits.maxCustomerOrganizations === -1 ? 'unlimited' : limits.maxCustomerOrganizations;
      case 'usersPerCustomer':
        return limits.maxUsersPerCustomerOrganization === -1 ? 'unlimited' : limits.maxUsersPerCustomerOrganization;
      case 'deploymentsPerCustomer':
        return limits.maxDeploymentsPerCustomerOrganization === -1
          ? 'unlimited'
          : limits.maxDeploymentsPerCustomerOrganization;
      default:
        return never(metric);
    }
  }

  getCurrentPlanLimit(
    metric: 'customerOrganizations' | 'usersPerCustomer' | 'deploymentsPerCustomer'
  ): string | number {
    const info = this.subscriptionInfo();
    if (!info) {
      return '';
    }
    return this.getPlanLimit(info.subscriptionType, metric);
  }

  canSelectStarterPlan(): boolean {
    const info = this.subscriptionInfo();
    if (!info) {
      return true;
    }

    // Check if current usage exceeds starter plan limits
    return (
      info.currentCustomerOrganizationCount <= info.starterLimits.maxCustomerOrganizations &&
      info.currentMaxUsersPerCustomer <= info.starterLimits.maxUsersPerCustomerOrganization &&
      info.currentMaxDeploymentTargetsPerCustomer <= info.starterLimits.maxDeploymentsPerCustomerOrganization
    );
  }

  getPlanDisplayName(subscriptionType: SubscriptionType): string {
    switch (subscriptionType) {
      case 'trial':
        return 'Trial';
      case 'starter':
        return 'Distr Starter';
      case 'pro':
        return 'Distr Pro';
      case 'enterprise':
        return 'Distr Enterprise';
      default:
        return never(subscriptionType);
    }
  }

  isTrialSubscription(): boolean {
    const info = this.subscriptionInfo();
    return info?.subscriptionType === 'trial';
  }

  hasActiveSubscription(): boolean {
    const info = this.subscriptionInfo();
    return info?.subscriptionType === 'starter' || info?.subscriptionType === 'pro';
  }

  async manageSubscription() {
    try {
      await this.subscriptionService.openBillingPortal();
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
  }
}
