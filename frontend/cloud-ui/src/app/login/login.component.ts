import {Component, inject, OnDestroy, OnInit} from '@angular/core';
import {AuthService} from '../services/auth.service';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {distinctUntilChanged, filter, firstValueFrom, lastValueFrom, map, Subject, takeUntil} from 'rxjs';
import {ActivatedRoute, Router, RouterLink} from '@angular/router';

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
        this.formGroup.patchValue({email: email});
      });
  }

  public ngOnDestroy(): void {
    this.destroyed$.next();
  }

  public async submit(): Promise<void> {
    if (this.formGroup.valid) {
      const value = this.formGroup.value;
      await lastValueFrom(this.auth.login(value.email!, value.password!));
      const targetRoute = value.email === 'pmig+customer@glasskube.com' ? '/deployments' : '/';
      await this.router.navigate([targetRoute]);
    }
  }
}
