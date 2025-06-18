import {Component, inject, OnDestroy, OnInit, signal} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute, Router, RouterLink} from '@angular/router';
import {distinctUntilChanged, filter, lastValueFrom, map, Subject, takeUntil} from 'rxjs';
import {getFormDisplayedError} from '../../util/errors';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {AuthService} from '../services/auth.service';
import {ToastService} from '../services/toast.service';
import {AsyncPipe} from '@angular/common';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faGithub, faGoogle, faMicrosoft} from '@fortawesome/free-brands-svg-icons';

@Component({
  selector: 'app-login',
  imports: [ReactiveFormsModule, RouterLink, AutotrimDirective, AsyncPipe, FaIconComponent],
  templateUrl: './login.component.html',
})
export class LoginComponent implements OnInit, OnDestroy {
  public readonly formGroup = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [Validators.required]),
  });
  loading = false;
  public errorMessage?: string;
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly destroyed$ = new Subject<void>();
  private readonly toast = inject(ToastService);
  readonly loginConfig$ = this.auth.loginConfig();
  readonly isJWTLogin = signal(false);

  public ngOnInit(): void {
    this.route.queryParams
      .pipe(
        map((params) => params['email']),
        filter((email) => email),
        distinctUntilChanged(),
        takeUntil(this.destroyed$)
      )
      .subscribe((email) => {
        this.formGroup.patchValue({email});
      });
    this.route.queryParams
      .pipe(
        map((params) => params['inviteSuccess']),
        filter((inviteSuccess) => inviteSuccess),
        distinctUntilChanged(),
        takeUntil(this.destroyed$)
      )
      .subscribe((inviteSuccess) => {
        if (inviteSuccess === 'true') {
          this.toast.success('Account activated successfully. You can now log in!');
        }
      });
    const reason = this.route.snapshot.queryParamMap.get('reason');
    switch (reason) {
      case 'password-reset':
        this.toast.success('Your password has been updated, you can now log in.');
        break;
      case 'session-expired':
        this.toast.success('You have been logged out because your session has expired.');
        break;
      case 'oidc-failed':
        this.toast.error('Login with this provider failed unexpectedly.');
        break;
    }
    const jwt = this.route.snapshot.queryParamMap.get('jwt');
    if (jwt) {
      this.isJWTLogin.set(true);
      this.auth.loginWithToken(jwt);
      window.location.href = '/';
    }
  }

  public ngOnDestroy(): void {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  public async submit(): Promise<void> {
    this.formGroup.markAllAsTouched();
    this.errorMessage = undefined;
    if (this.formGroup.valid) {
      this.loading = true;
      const value = this.formGroup.value;
      try {
        await lastValueFrom(this.auth.login(value.email!, value.password!));
        if (this.auth.hasRole('customer')) {
          await this.router.navigate(['/home']);
        } else {
          await this.router.navigate(['/dashboard'], {queryParams: {from: 'login'}});
        }
      } catch (e) {
        this.errorMessage = getFormDisplayedError(e);
      } finally {
        this.loading = false;
      }
    }
  }

  protected getLoginURL(provider: string): string {
    return `/api/v1/auth/oidc/${provider}`;
  }

  protected readonly faGoogle = faGoogle;
  protected readonly faGithub = faGithub;
  protected readonly faMicrosoft = faMicrosoft;
}
