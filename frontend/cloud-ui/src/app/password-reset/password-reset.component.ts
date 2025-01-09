import {HttpErrorResponse} from '@angular/common/http';
import {Component, inject} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {lastValueFrom} from 'rxjs';
import {AuthService} from '../services/auth.service';
import {SettingsService} from '../services/settings.service';
import {getFormDisplayedError} from '../../util/errors';

@Component({
  selector: 'app-password-reset',
  imports: [ReactiveFormsModule],
  templateUrl: './password-reset.component.html',
})
export class PasswordResetComponent {
  private readonly settings = inject(SettingsService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  public readonly form = new FormGroup(
    {
      password: new FormControl('', [Validators.required, Validators.minLength(8)]),
      passwordConfirm: new FormControl('', [Validators.required]),
    },
    (control) => (control.value.password === control.value.passwordConfirm ? null : {passwordMismatch: 'error'})
  );
  public readonly email = this.auth.getClaims().email;
  public errorMessage?: string;
  loading = false;

  public async submit() {
    this.form.markAllAsTouched();
    this.errorMessage = undefined;
    if (this.form.valid) {
      this.loading = true;
      try {
        await lastValueFrom(this.settings.updateUserSettings({password: this.form.value.password!}));
        await lastValueFrom(this.auth.logout());
        location.assign(`/login?email=${encodeURIComponent(this.email)}`);
      } catch (e) {
        this.errorMessage = getFormDisplayedError(e);
        // TODO maybe check for 429 again and disable the button for some time? (but how long then?)
      } finally {
        this.loading = false;
      }
    }
  }
}
