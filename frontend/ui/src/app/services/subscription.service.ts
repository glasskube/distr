import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {firstValueFrom, Observable} from 'rxjs';
import {CheckoutRequest, SubscriptionInfo} from '../types/subscription';

@Injectable({
  providedIn: 'root',
})
export class SubscriptionService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/billing/subscription';

  get(): Observable<SubscriptionInfo> {
    return this.httpClient.get<SubscriptionInfo>(this.baseUrl);
  }

  async checkout(request: CheckoutRequest): Promise<void> {
    console.log('checkout');

    // Create checkout session on backend
    const response = await firstValueFrom(
      this.httpClient.post<{
        sessionId: string;
        url: string;
      }>(this.baseUrl, request)
    );

    if (!response?.url) {
      throw new Error('Failed to create checkout session');
    }

    window.location.href = response.url;
  }
}
