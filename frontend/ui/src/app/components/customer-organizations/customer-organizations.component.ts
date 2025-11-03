import {AsyncPipe, DatePipe, DecimalPipe} from '@angular/common';
import {Component, inject, TemplateRef, viewChild, ViewChild} from '@angular/core';
import {toSignal} from '@angular/core/rxjs-interop';
import {FormBuilder, FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FontAwesomeModule} from '@fortawesome/angular-fontawesome';
import {
  faCircleExclamation,
  faMagnifyingGlass,
  faPlus,
  faTrash,
  faUserCircle,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, filter, firstValueFrom, map, startWith, Subject, switchMap} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {SecureImagePipe} from '../../../util/secureImage';
import {modalFlyInOut} from '../../animations/modal';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {CustomerOrganizationsService} from '../../services/customer-organizations.service';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {CustomerOrganization} from '../../types/customer-organization';
import {UuidComponent} from '../uuid';
import {RouterLink} from '@angular/router';

@Component({
  templateUrl: './customer-organizations.component.html',
  imports: [
    ReactiveFormsModule,
    FontAwesomeModule,
    UuidComponent,
    DatePipe,
    RequireRoleDirective,
    SecureImagePipe,
    AsyncPipe,
    DecimalPipe,
    RouterLink,
  ],
  animations: [modalFlyInOut],
})
export class CustomerOrganizationsComponent {
  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faPlus = faPlus;
  protected readonly faUserCircle = faUserCircle;
  protected readonly faTrash = faTrash;
  protected readonly faXmark = faXmark;
  protected readonly faCircleExclamation = faCircleExclamation;

  private readonly customerOrganizationsService = inject(CustomerOrganizationsService);
  private readonly toast = inject(ToastService);
  private readonly overlay = inject(OverlayService);
  private readonly fb = inject(FormBuilder).nonNullable;
  protected readonly featureFlags = inject(FeatureFlagService);

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
  protected readonly createForm = new FormGroup({
    name: new FormControl('', [Validators.required]),
  });
  protected createFormLoading = false;

  protected showCreateDialog() {
    this.closeCreateDialog();
    this.modalRef = this.overlay.showModal(this.createCustomerDialog());
  }

  protected closeCreateDialog(reset: boolean = true): void {
    this.modalRef?.close();

    if (reset) {
      this.createForm.reset();
    }
  }

  protected submitCreateForm() {
    this.createForm.markAllAsTouched();

    if (this.createForm.invalid) {
      return;
    }

    this.createFormLoading = true;
    this.customerOrganizationsService
      .createCustomerOrganization({
        name: this.createForm.value.name!,
      })
      .subscribe({
        next: () => {
          this.createFormLoading = false;
          this.closeCreateDialog();
          this.refresh$.next();
        },
        error: () => {
          this.createFormLoading = false;
        },
      });
  }

  protected async uploadImage(target: CustomerOrganization): Promise<void> {
    const fileId = await firstValueFrom(this.overlay.uploadImage({scope: 'platform'}));
    if (!fileId || fileId == target.imageId) {
      return;
    }
    target.imageId = fileId;
    await firstValueFrom(this.customerOrganizationsService.updateCustomerOrganization(target));
    this.refresh$.next();
  }

  protected delete(target: CustomerOrganization) {
    this.overlay
      .confirm({message: {message: 'Are you sure you want to delete this customer organization?'}})
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
