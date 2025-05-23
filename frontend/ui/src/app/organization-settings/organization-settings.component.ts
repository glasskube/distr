import {Component, inject, OnInit, signal} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faFloppyDisk, faLightbulb} from '@fortawesome/free-solid-svg-icons';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {firstValueFrom, lastValueFrom} from 'rxjs';
import {getFormDisplayedError} from '../../util/errors';
import {ToastService} from '../services/toast.service';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {OrganizationService} from '../services/organization.service';
import {Organization} from '../types/organization';
import {slugMaxLength, slugPattern} from '../../util/slug';

@Component({
  selector: 'app-organization-settings',
  templateUrl: './organization-settings.component.html',
  imports: [FaIconComponent, ReactiveFormsModule, AutotrimDirective],
})
export class OrganizationSettingsComponent implements OnInit {
  protected readonly faFloppyDisk = faFloppyDisk;
  protected readonly faLightbulb = faLightbulb;

  private readonly organizationService = inject(OrganizationService);
  private organization?: Organization;
  private toast = inject(ToastService);

  protected readonly form = new FormGroup({
    name: new FormControl('', [Validators.required]),
    slug: new FormControl('', [Validators.pattern(slugPattern), Validators.maxLength(slugMaxLength)]),
    appDomain: new FormControl<string | undefined>({value: undefined, disabled: true}),
    registryDomain: new FormControl<string | undefined>({value: undefined, disabled: true}),
    emailFromAddress: new FormControl<string | undefined>({value: undefined, disabled: true}),
  });
  formLoading = signal(false);

  async ngOnInit() {
    try {
      this.organization = await firstValueFrom(this.organizationService.get());
      if (this.organization.slug) {
        this.form.controls.slug.addValidators([Validators.required]);
      }
      this.form.patchValue(this.organization);
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
  }

  async save() {
    this.form.markAllAsTouched();
    if (this.form.valid) {
      this.formLoading.set(true);
      try {
        this.organization = await lastValueFrom(
          this.organizationService.update({
            ...this.organization!,
            name: this.form.value.name?.trim(),
            slug: this.form.value.slug?.trim(),
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
