import {Component, inject} from '@angular/core';
import {SettingsService} from '../services/settings.service';
import {firstValueFrom} from 'rxjs';
import {ToastService} from '../services/toast.service';
import {AuthService} from '../services/auth.service';

@Component({
  selector: 'app-verify',
  templateUrl: './verify.component.html',
  imports: [],
})
export class VerifyComponent {
  private readonly settings = inject(SettingsService);
  private readonly toast = inject(ToastService);
  private readonly auth = inject(AuthService);
  public requestMailEnabled = true;

  public async requestMail() {
    this.requestMailEnabled = false;
    try {
      await firstValueFrom(this.settings.requestEmailVerification());
      this.toast.success('Verification email has been sent. Check your inbox.');
    } catch (e) {
      this.requestMailEnabled = true;
    }
  }

  public async logoutAndRedirectToLogin() {
    await firstValueFrom(this.auth.logout());
    location.assign('/login');
  }
}
