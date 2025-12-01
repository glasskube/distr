import {Component, inject, OnInit, signal} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faShoppingCart, faCreditCard} from '@fortawesome/free-solid-svg-icons';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {firstValueFrom} from 'rxjs';
import {getFormDisplayedError} from '../../util/errors';
import {ToastService} from '../services/toast.service';
import {SubscriptionService} from '../services/subscription.service';
import {SubscriptionInfo, SubscriptionType} from '../types/subscription';
import {CommonModule} from '@angular/common';
import {never} from '../../util/exhaust';

@Component({
  selector: 'app-subscription',
  templateUrl: './subscription.component.html',
  imports: [FaIconComponent, ReactiveFormsModule, CommonModule],
})
export class SubscriptionComponent implements OnInit {
  protected readonly faShoppingCart = faShoppingCart;
  protected readonly faCreditCard = faCreditCard;

  private readonly subscriptionService = inject(SubscriptionService);
  private readonly toast = inject(ToastService);

  protected subscriptionInfo = signal<SubscriptionInfo | undefined>(undefined);

  protected readonly form = new FormGroup({
    subscriptionType: new FormControl<SubscriptionType>('pro', [Validators.required]),
    billingMode: new FormControl<'monthly' | 'yearly'>('monthly', [Validators.required]),
    userAccountQuantity: new FormControl<number>(1, [Validators.required, Validators.min(1)]),
    customerOrganizationQuantity: new FormControl<number>(1, [Validators.required, Validators.min(1)]),
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
      try {
        const body = {
          subscriptionUserAccountQuantity: this.form.value.userAccountQuantity!,
          subscriptionCustomerOrganizationQuantity: this.form.value.customerOrganizationQuantity!,
        };

        // Call the update subscription endpoint
        const updatedInfo = await this.subscriptionService.updateSubscription(body);

        // Update the subscription info signal with the new data
        this.subscriptionInfo.set(updatedInfo);

        // Show success message
        this.toast.success('Subscription updated successfully');
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      }
    }
  }

  getPlanLimits(plan: SubscriptionType): {customers: string; users: string; deployments: string} {
    const info = this.subscriptionInfo();
    if (!info) {
      return {customers: '', users: '', deployments: ''};
    }

    let limits;
    switch (plan) {
      case 'trial':
        limits = info.trialLimits;
        break;
      case 'starter':
        limits = info.starterLimits;
        break;
      case 'pro':
        limits = info.proLimits;
        break;
      case 'enterprise':
        limits = info.enterpriseLimits;
        break;
      default:
        return never(plan);
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

  getCurrentPlanLimit(
    metric: 'customerOrganizations' | 'usersPerCustomer' | 'deploymentsPerCustomer'
  ): string | number {
    const info = this.subscriptionInfo();
    if (!info) {
      return '';
    }

    let limits;
    switch (info.subscriptionType) {
      case 'trial':
        limits = info.trialLimits;
        break;
      case 'starter':
        limits = info.starterLimits;
        break;
      case 'pro':
        limits = info.proLimits;
        break;
      case 'enterprise':
        limits = info.enterpriseLimits;
        break;
      default:
        return never(info.subscriptionType);
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
