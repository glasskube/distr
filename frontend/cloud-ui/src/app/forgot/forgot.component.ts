import {Component, inject, OnDestroy, OnInit} from '@angular/core';
import {FormGroup, FormControl, Validators, ReactiveFormsModule} from '@angular/forms';
import {Router, ActivatedRoute, RouterLink} from '@angular/router';
import {distinctUntilChanged, filter, lastValueFrom, map, Subject, takeUntil} from 'rxjs';
import {AuthService} from '../services/auth.service';
import {HttpErrorResponse} from '@angular/common/http';

@Component({
  selector: 'app-forgot',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './forgot.component.html',
})
export class ForgotComponent implements OnInit, OnDestroy {
  public readonly formGroup = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
  });
  public errorMessage?: string;
  public success = false;
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly destroyed$ = new Subject<void>();

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
  }

  public ngOnDestroy(): void {
    this.destroyed$.next();
  }

  public async submit(): Promise<void> {
    this.formGroup.markAllAsTouched();
    this.errorMessage = undefined;
    if (this.formGroup.valid) {
      const value = this.formGroup.value;
      try {
        await lastValueFrom(this.auth.resetPassword(value.email!));
      } catch (e) {
        if (e instanceof HttpErrorResponse && e.status < 500 && typeof e.error === 'string') {
          this.errorMessage = e.error;
        } else {
          console.error(e);
          this.errorMessage = 'something went wrong';
        }
        return;
      }
      this.success = true;
    }
  }
}
