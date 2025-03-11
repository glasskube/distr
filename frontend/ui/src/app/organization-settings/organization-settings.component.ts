import {Component, inject, OnInit, signal} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faFloppyDisk} from '@fortawesome/free-solid-svg-icons';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {firstValueFrom} from 'rxjs';
import {HttpErrorResponse} from '@angular/common/http';
import {getFormDisplayedError} from '../../util/errors';
import {ToastService} from '../services/toast.service';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {OrganizationService} from '../services/organization.service';
import {Organization} from '../types/organization';

@Component({
  selector: 'app-organization-settings',
  templateUrl: './organization-settings.component.html',
  imports: [FaIconComponent, ReactiveFormsModule, AutotrimDirective],
})
export class OrganizationSettingsComponent implements OnInit {
  protected readonly faFloppyDisk = faFloppyDisk;

  private readonly organizationService = inject(OrganizationService);
  private organization?: Organization;
  private toast = inject(ToastService);

  readonly slugPattern = /^[a-z]+$/;

  protected readonly form = new FormGroup({
    name: new FormControl('', [Validators.required]),
    slug: new FormControl('', [Validators.pattern(this.slugPattern)]),
  });
  formLoading = signal(false);

  async ngOnInit() {
    try {
      this.organization = await firstValueFrom(this.organizationService.get());
      if (this.organization.slug) {
        this.form.controls.slug.addValidators([Validators.required]);
      }
      this.form.patchValue({
        name: this.organization.name,
        slug: this.organization.slug,
      });
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg && e instanceof HttpErrorResponse && e.status !== 404) {
        // it's a valid use case for an organization to have no branding (hence 404 is not shown in toast)
        this.toast.error(msg);
      }
    }
  }

  async save() {
    this.form.markAllAsTouched();
    if (this.form.valid) {
      this.formLoading.set(true);
      try {
        this.organization = await firstValueFrom(
          this.organizationService.update({
            ...this.organization!,
            name: this.form.value.name!.trim(),
            slug: this.form.value.slug!.trim(),
          })
        );
        this.toast.success('Settings saved successfully');
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.formLoading.set(false);
      }
    }
  }
}
