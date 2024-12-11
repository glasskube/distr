import {Component, inject, OnDestroy, OnInit} from '@angular/core';
import {AuthService} from '../services/auth.service';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {distinctUntilChanged, filter, firstValueFrom, lastValueFrom, map, Subject, takeUntil} from 'rxjs';
import {ActivatedRoute, Router, RouterLink} from '@angular/router';

@Component({
  selector: 'app-verify',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './verify.component.html',
})
export class VerifyComponent implements OnInit, OnDestroy {
  public readonly formGroup = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [Validators.required]),
  });
  private readonly route = inject(ActivatedRoute);
  private readonly destroyed$ = new Subject<void>();

  public ngOnInit(): void {
    this.route.queryParams
      .pipe(
        map((params) => params['jwt']),
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

}
