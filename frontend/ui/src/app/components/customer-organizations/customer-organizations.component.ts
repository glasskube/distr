import {AsyncPipe, DatePipe, DecimalPipe} from '@angular/common';
import {Component, computed, inject, TemplateRef, viewChild} from '@angular/core';
import {toSignal} from '@angular/core/rxjs-interop';
import {FormBuilder, ReactiveFormsModule, Validators} from '@angular/forms';
import {RouterLink} from '@angular/router';
import {CustomerOrganization} from '@distr-sh/distr-sdk';
import {FontAwesomeModule} from '@fortawesome/angular-fontawesome';
import {
  faBuildingUser,
  faCircleExclamation,
  faEdit,
  faMagnifyingGlass,
  faPlus,
  faTrash,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, filter, firstValueFrom, map, startWith, Subject, switchMap} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {SecureImagePipe} from '../../../util/secureImage';
import {modalFlyInOut} from '../../animations/modal';
import {RequireVendorDirective} from '../../directives/required-role.directive';
import {AuthService} from '../../services/auth.service';
import {CustomerOrganizationsService} from '../../services/customer-organizations.service';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {ImageUploadService} from '../../services/image-upload.service';
import {OrganizationService} from '../../services/organization.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {QuotaLimitComponent} from '../quota-limit.component';
import {UuidComponent} from '../uuid';

@Component({
  templateUrl: './customer-organizations.component.html',
  imports: [
    ReactiveFormsModule,
    FontAwesomeModule,
    UuidComponent,
    DatePipe,
    SecureImagePipe,
    AsyncPipe,
    DecimalPipe,
    RouterLink,
    RequireVendorDirective,
    QuotaLimitComponent,
  ],
  animations: [modalFlyInOut],
})
export class CustomerOrganizationsComponent {
  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faPlus = faPlus;
  protected readonly faBuildingUser = faBuildingUser;
  protected readonly faTrash = faTrash;
  protected readonly faXmark = faXmark;
  protected readonly faCircleExclamation = faCircleExclamation;
  protected readonly faEdit = faEdit;

  private readonly customerOrganizationsService = inject(CustomerOrganizationsService);
  private readonly toast = inject(ToastService);
  private readonly imageUploadService = inject(ImageUploadService);
  private readonly overlay = inject(OverlayService);
  private readonly fb = inject(FormBuilder).nonNullable;
  private readonly organizationService = inject(OrganizationService);
  protected readonly featureFlags = inject(FeatureFlagService);
  protected readonly auth = inject(AuthService);

  private readonly organization = toSignal(this.organizationService.get());
  protected readonly limit = computed(() => this.organization()?.subscriptionCustomerOrganizationQuantity);

  protected readonly filterForm = this.fb.group({
    search: this.fb.control(''),
  });
  private readonly refresh$ = new Subject<void>();
  protected readonly customerOrganizations = toSignal(
    combineLatest([
      this.filterForm.valueChanges.pipe(
        map((filter) => filter.search ?? ''),
        startWith('')
      ),
      this.refresh$.pipe(
        startWith(undefined),
        switchMap(() => this.customerOrganizationsService.getCustomerOrganizations())
      ),
    ]).pipe(
      map(([filter, organizations]) =>
        filter.length > 0
          ? organizations.filter((organization) => organization.name.toLowerCase().includes(filter.toLowerCase()))
          : organizations
      )
    )
  );

  private readonly createCustomerDialog = viewChild.required<TemplateRef<unknown>>('createCustomerDialog');
  private modalRef?: DialogRef;
  protected readonly createForm = this.fb.group({
    id: this.fb.control(''),
    name: this.fb.control('', [Validators.required]),
    imageId: this.fb.control(''),
  });
  protected createFormLoading = false;

  protected showCreateDialog() {
    this.closeCreateDialog();
    this.modalRef = this.overlay.showModal(this.createCustomerDialog());
  }

  protected showUpdateDialog(value: CustomerOrganization) {
    this.closeCreateDialog();
    this.createForm.patchValue(value);
    this.modalRef = this.overlay.showModal(this.createCustomerDialog());
  }

  protected closeCreateDialog(reset: boolean = true): void {
    this.modalRef?.close();

    if (reset) {
      this.createForm.reset();
    }
  }

  protected async submitCreateForm() {
    this.createForm.markAllAsTouched();

    if (this.createForm.invalid) {
      return;
    }

    this.createFormLoading = true;

    const request = {
      name: this.createForm.value.name!,
      imageId: this.createForm.value.imageId || undefined,
    };

    try {
      if (this.createForm.value.id) {
        await firstValueFrom(
          this.customerOrganizationsService.updateCustomerOrganization(this.createForm.value.id, request)
        );
      } else {
        await firstValueFrom(this.customerOrganizationsService.createCustomerOrganization(request));
      }

      this.closeCreateDialog();
      this.refresh$.next();
    } finally {
      this.createFormLoading = false;
    }
  }

  protected async uploadImage(value: CustomerOrganization): Promise<void> {
    const imageId = await firstValueFrom(this.imageUploadService.showDialog({scope: 'platform'}));
    if (!imageId || imageId === value.imageId) {
      return;
    }
    await firstValueFrom(
      this.customerOrganizationsService.updateCustomerOrganization(value.id, {name: value.name, imageId})
    );
    this.refresh$.next();
  }

  protected delete(target: CustomerOrganization) {
    this.overlay
      .confirm({message: {message: 'Are you sure you want to delete this customer?'}})
      .pipe(
        filter((it) => it === true),
        switchMap(() => this.customerOrganizationsService.deleteCustomerOrganization(target.id!))
      )
      .subscribe({
        next: () => this.refresh$.next(),
        error: (e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
        },
      });
  }
}
