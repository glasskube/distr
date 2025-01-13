import {Component, inject} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faFloppyDisk} from '@fortawesome/free-solid-svg-icons';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {OrganizationBrandingService} from '../services/organization-branding.service';
import {lastValueFrom} from 'rxjs';

@Component({
  selector: 'app-organization-branding',
  templateUrl: './organization-branding.component.html',
  imports: [FaIconComponent, ReactiveFormsModule],
})
export class OrganizationBrandingComponent {
  protected readonly faFloppyDisk = faFloppyDisk;

  private readonly organizationBranding = inject(OrganizationBrandingService);
  private readonly organizationBranding$ = this.organizationBranding.get();

  protected readonly form = new FormGroup({
    title: new FormControl('', Validators.required),
    description: new FormControl('', Validators.required),
    logo: new FormControl<Blob | null>(null),
  });

  constructor() {
    this.organizationBranding$.subscribe((branding) => {
      this.form.patchValue({
        title: branding.title,
        description: branding.description,
        logo: this.base64ToBlob(branding.logo!!, branding.logoContentType),
      });
    });
  }

  async save() {
    this.form.markAllAsTouched();
    if (this.form.valid) {
      const formData = new FormData();

      const id = (await lastValueFrom(this.organizationBranding$))?.id;

      if (this.form.value.logo) {
        formData.set('title', this.form.value.title!!);
        formData.set('description', this.form.value.description!!);
        formData.set('logo', this.form.value.logo as File);
      }

      if (id) {
        formData.set('id', id);
        await lastValueFrom(this.organizationBranding.update(formData));
      } else {
        await lastValueFrom(this.organizationBranding.create(formData));
      }
    }
  }

  onLogoChange(event: any) {
    const file = (event.target as HTMLInputElement).files?.[0];
    this.form.patchValue({logo: file ?? null});
  }

  deleteLogo() {
    this.form.patchValue({logo: null});
  }

  getLogoString(): string | null {
    return this.form.value.logo ? URL.createObjectURL(this.form.value.logo) : null;
  }

  private base64ToBlob(base64String: string, contentType = ''): Blob {
    const byteCharacters = atob(base64String);
    const byteArrays = [];

    for (let i = 0; i < byteCharacters.length; i++) {
      byteArrays.push(byteCharacters.charCodeAt(i));
    }

    const byteArray = new Uint8Array(byteArrays);
    return new Blob([byteArray], {type: contentType});
  }
}
