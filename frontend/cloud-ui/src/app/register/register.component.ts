import {Component, inject} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router, RouterLink} from '@angular/router';
import {firstValueFrom} from 'rxjs';
import {AuthService} from '../services/auth.service';

@Component({
  selector: 'app-register',
  imports: [RouterLink, ReactiveFormsModule],
  templateUrl: './register.component.html',
})
export class RegisterComponent {
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);

  public readonly form = new FormGroup(
    {
      email: new FormControl('', [Validators.required, Validators.email]),
      name: new FormControl<string | undefined>(undefined),
      password: new FormControl('', [Validators.required, Validators.minLength(8)]),
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
      await firstValueFrom(this.auth.register(value.email!, value.name, value.password!));
      await this.router.navigate(['/login'], {queryParams: {email: value.email!}});
    }
  }
}
