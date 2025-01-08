import {Component, inject} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {firstValueFrom} from 'rxjs';
import {AuthService} from '../services/auth.service';
import {SettingsService} from '../services/settings.service';

@Component({
  selector: 'app-invite',
  imports: [ReactiveFormsModule],
  templateUrl: './invite.component.html',
})
export class InviteComponent {
  private readonly auth = inject(AuthService);
  private readonly settings = inject(SettingsService);
  public readonly email = this.auth.getClaims().email;

  public readonly form = new FormGroup(
    {
      name: new FormControl<string | undefined>(this.auth.getClaims().name, {nonNullable: true}),
      password: new FormControl('', {nonNullable: true, validators: [Validators.required, Validators.minLength(8)]}),
      passwordConfirm: new FormControl('', [Validators.required]),
      terms: new FormControl(false, Validators.required),
    },
    (control) => (control.value.password === control.value.passwordConfirm ? null : {passwordMismatch: 'error'})
  );
  public submitted = false;

  public async submit(): Promise<void> {
    if (this.form.valid) {
      this.submitted = true;
      const value = this.form.value;
      await firstValueFrom(this.settings.updateUserSettings({name: value.name, password: value.password}));
      if (this.auth.getClaims().email_verified) {
        await firstValueFrom(this.settings.confirmEmailVerification());
      }
      this.auth.logout();
      location.assign(`/login?email=${encodeURIComponent(this.email)}`);
    }
  }
}
