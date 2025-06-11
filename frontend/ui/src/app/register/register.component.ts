import {Component, inject, OnInit} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute, Router, RouterLink} from '@angular/router';
import {firstValueFrom} from 'rxjs';
import {getFormDisplayedError} from '../../util/errors';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {AuthService} from '../services/auth.service';
import {ToastService} from '../services/toast.service';

@Component({
  selector: 'app-register',
  imports: [RouterLink, ReactiveFormsModule, AutotrimDirective],
  templateUrl: './register.component.html',
})
export class RegisterComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly toast = inject(ToastService);
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);

  errorMessage?: string;
  loading = false;
  public readonly form = new FormGroup(
    {
      email: new FormControl('', [Validators.required, Validators.email]),
      name: new FormControl<string | undefined>(undefined),
      password: new FormControl('', [Validators.required, Validators.minLength(8)]),
      passwordConfirm: new FormControl('', [Validators.required]),
    },
    (control) => (control.value.password === control.value.passwordConfirm ? null : {passwordMismatch: 'error'})
  );

  ngOnInit() {
    const reason = this.route.snapshot.queryParamMap.get('reason');
    switch(reason) {
      case 'oidc-user-not-found':
        this.toast.error('This email address is not associated with a Distr account yet. Please create it here first.');
        break;
    }
    const email = this.route.snapshot.queryParamMap.get('email');
    if(email) {
      this.form.patchValue({email});
    }
  }

  public async submit(): Promise<void> {
    this.form.markAllAsTouched();
    this.errorMessage = undefined;
    if (this.form.valid) {
      this.loading = true;
      const value = this.form.value;
      try {
        await firstValueFrom(this.auth.register(value.email!, value.name, value.password!));
        await this.router.navigate(['/login'], {queryParams: {email: value.email!}});
      } catch (e) {
        this.errorMessage = getFormDisplayedError(e);
      } finally {
        this.loading = false;
      }
    }
  }
}
