import {Component, inject, OnInit, signal} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faShoppingCart} from '@fortawesome/free-solid-svg-icons';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {firstValueFrom} from 'rxjs';
import {getFormDisplayedError} from '../../util/errors';
import {ToastService} from '../services/toast.service';
import {SubscriptionService} from '../services/subscription.service';
import {SubscriptionInfo, SubscriptionType} from '../types/subscription';
import {CommonModule} from '@angular/common';

@Component({
  selector: 'app-subscription',
  templateUrl: './subscription.component.html',
  imports: [FaIconComponent, ReactiveFormsModule, CommonModule],
})
export class SubscriptionComponent implements OnInit {
  protected readonly faShoppingCart = faShoppingCart;

  private readonly subscriptionService = inject(SubscriptionService);
  private readonly toast = inject(ToastService);

  protected subscriptionInfo = signal<SubscriptionInfo | null>(null);
  protected loading = signal(true);

  protected readonly form = new FormGroup({
    subscriptionType: new FormControl<SubscriptionType>('starter', [Validators.required]),
    billingMode: new FormControl<'monthly' | 'yearly'>('monthly', [Validators.required]),
    userAccountQuantity: new FormControl<number>(1, [Validators.required, Validators.min(1)]),
    customerOrganizationQuantity: new FormControl<number>(1, [Validators.required, Validators.min(1)]),
    currency: new FormControl<string>('usd', [Validators.required]),
  });

  async ngOnInit() {
    try {
      const info = await firstValueFrom(this.subscriptionService.get());
      this.subscriptionInfo.set(info);

      // Pre-fill form with current subscription values or defaults
      this.form.patchValue({
        subscriptionType: info.subscriptionType === 'trial' ? 'starter' : info.subscriptionType,
        userAccountQuantity: info.subscriptionUserAccountQuantity ?? info.currentUserAccountCount,
        customerOrganizationQuantity:
          info.subscriptionCustomerOrganizationQuantity ?? info.currentCustomerOrganizationCount,
      });
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    } finally {
      this.loading.set(false);
    }
  }

  getPreviewPrice(): number {
    const subscriptionType = this.form.value.subscriptionType;
    const billingMode = this.form.value.billingMode;
    const userQty = this.form.value.userAccountQuantity ?? 0;
    const customerQty = this.form.value.customerOrganizationQuantity ?? 0;

    let basePrice = 0;
    let userPrice = 0;
    let customerPrice = 0;

    if (subscriptionType === 'starter') {
      basePrice = billingMode === 'monthly' ? 40 : 480;
      userPrice = billingMode === 'monthly' ? 16 : 192;
      customerPrice = billingMode === 'monthly' ? 24 : 288;
    } else if (subscriptionType === 'pro') {
      basePrice = billingMode === 'monthly' ? 80 : 960;
      userPrice = billingMode === 'monthly' ? 24 : 288;
      customerPrice = billingMode === 'monthly' ? 56 : 672;
    }

    return basePrice + userPrice * userQty + customerPrice * customerQty;
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
          currency: this.form.value.currency!,
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

  getPlanLimits(plan: SubscriptionType): {customers: string; users: string; deployments: string} {
    if (plan === 'starter') {
      return {customers: 'Up to 3', users: '1 per customer', deployments: '1 per customer'};
    } else if (plan === 'pro') {
      return {customers: 'Up to 100', users: 'Up to 10 per customer', deployments: '3 per customer'};
    }
    return {customers: 'Unlimited', users: 'Unlimited', deployments: 'Unlimited'};
  }

  getCurrencySymbol(): string {
    const currency = this.form.value.currency || 'usd';
    return currency === 'eur' ? '€' : '$';
  }
}
