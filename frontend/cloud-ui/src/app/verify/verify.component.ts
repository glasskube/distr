import {Component, inject} from '@angular/core';
import {SettingsService} from '../services/settings.service';
import {firstValueFrom, take, timeout, timer} from 'rxjs';
import {ToastService} from '../services/toast.service';

@Component({
  selector: 'app-verify',
  templateUrl: './verify.component.html',
})
export class VerifyComponent {
  private readonly settings = inject(SettingsService);
  private readonly toast = inject(ToastService);
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
}
