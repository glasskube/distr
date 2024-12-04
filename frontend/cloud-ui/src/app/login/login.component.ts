import {Component, inject} from '@angular/core';
import {AuthService} from '../services/auth.service';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {firstValueFrom, lastValueFrom} from 'rxjs';
import {Router} from '@angular/router';

@Component({
  selector: 'app-login',
  imports: [ReactiveFormsModule],
  templateUrl: './login.component.html',
})
export class LoginComponent {
  public readonly formGroup = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [Validators.required]),
  });
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  public async submit(): Promise<void> {
    if (this.formGroup.valid) {
      const value = this.formGroup.value;
      await lastValueFrom(this.auth.login(value.email!, value.password!));
      await this.router.navigate(['/']);
    }
  }
}
