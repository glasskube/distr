import {Component, inject, OnDestroy, OnInit} from '@angular/core';
import {AuthService} from '../services/auth.service';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {distinctUntilChanged, filter, firstValueFrom, lastValueFrom, map, Subject, takeUntil} from 'rxjs';
import {ActivatedRoute, Router, RouterLink} from '@angular/router';
import {SettingsService} from '../services/settings.service';
import {ToastService} from '../services/toast.service';

@Component({
  selector: 'app-verify',
  imports: [ReactiveFormsModule],
  templateUrl: './verify.component.html',
})
export class VerifyComponent implements OnInit, OnDestroy {
  private readonly auth = inject(AuthService);
  private readonly settings = inject(SettingsService);
  private readonly toast = inject(ToastService);
  public readonly formGroup = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [Validators.required]),
  });
  private readonly router = inject(Router);
  private readonly destroyed$ = new Subject<void>();

  async ngOnInit() {
    if(this.auth.getClaims().email_verified) {
      await firstValueFrom(this.settings.updateUserSettings({emailVerified: true}));
      this.toast.success("Your email has been verified");
      await this.router.navigate(['/']);
    }
  }

  public ngOnDestroy(): void {
    this.destroyed$.next();
  }

}
