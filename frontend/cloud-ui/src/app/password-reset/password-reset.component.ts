import {HttpErrorResponse} from '@angular/common/http';
import {Component, inject} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {lastValueFrom} from 'rxjs';
import {AuthService} from '../services/auth.service';
import {SettingsService} from '../services/settings.service';

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

  public async submit() {
    this.form.markAllAsTouched();
    this.errorMessage = undefined;
    if (this.form.valid) {
      try {
        await lastValueFrom(this.settings.updateUserSettings({password: this.form.value.password!}));
      } catch (e) {
        if (e instanceof HttpErrorResponse && e.status < 500 && typeof e.error === 'string') {
          this.errorMessage = e.error;
        } else {
          this.errorMessage = 'something went wrong';
        }
        console.error(e);
      }
      await lastValueFrom(this.auth.logout());
      location.assign(`/login?email=${encodeURIComponent(this.email)}`);
    }
  }
}
