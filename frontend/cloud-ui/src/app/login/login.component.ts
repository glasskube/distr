import {Component, inject, OnDestroy, OnInit} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute, Router, RouterLink} from '@angular/router';
import {distinctUntilChanged, filter, lastValueFrom, map, Subject, takeUntil} from 'rxjs';
import {AuthService} from '../services/auth.service';
import {HttpErrorResponse} from '@angular/common/http';
import {getFormDisplayedError} from '../../util/errors';
import {ToastService} from '../services/toast.service';

@Component({
  selector: 'app-login',
  imports: [ReactiveFormsModule, RouterLink],
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
          await this.router.navigate(['/dashboard']);
        }
      } catch (e) {
        this.errorMessage = getFormDisplayedError(e);
      } finally {
        this.loading = false;
      }
    }
  }
}
