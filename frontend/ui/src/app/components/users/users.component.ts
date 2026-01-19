import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, computed, inject, input, output, signal, TemplateRef, viewChild} from '@angular/core';
import {toObservable, toSignal} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {UserAccountWithRole, UserRole} from '@distr-sh/distr-sdk';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faBox,
  faCheck,
  faCircleExclamation,
  faClipboard,
  faMagnifyingGlass,
  faPen,
  faPlus,
  faRepeat,
  faTrash,
  faUserCircle,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {catchError, filter, firstValueFrom, NEVER, switchMap, tap} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {filteredByFormControl} from '../../../util/filter';
import {SecureImagePipe} from '../../../util/secureImage';
import {modalFlyInOut} from '../../animations/modal';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {RequireVendorDirective} from '../../directives/required-role.directive';
import {AuthService} from '../../services/auth.service';
import {ImageUploadService} from '../../services/image-upload.service';
import {OrganizationService} from '../../services/organization.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {UsersService} from '../../services/users.service';
import {QuotaLimitComponent} from '../quota-limit.component';
import {UuidComponent} from '../uuid';

@Component({
  selector: 'app-users',
  imports: [
    FaIconComponent,
    AsyncPipe,
    DatePipe,
    ReactiveFormsModule,
    RequireVendorDirective,
    AutotrimDirective,
    UuidComponent,
    SecureImagePipe,
    QuotaLimitComponent,
  ],
  templateUrl: './users.component.html',
  animations: [modalFlyInOut],
})
export class UsersComponent {
  public readonly users = input.required<UserAccountWithRole[]>();
  public readonly customerOrganizationId = input<string>();
  public readonly refresh = output<void>();

  private readonly toast = inject(ToastService);
  private readonly usersService = inject(UsersService);
  private readonly organizationService = inject(OrganizationService);
  private readonly overlay = inject(OverlayService);
  private readonly imageUploadService = inject(ImageUploadService);
  protected readonly auth = inject(AuthService);

  protected readonly faBox = faBox;
  protected readonly faCheck = faCheck;
  protected readonly faCircleExclamation = faCircleExclamation;
  protected readonly faClipboard = faClipboard;
  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faPen = faPen;
  protected readonly faPlus = faPlus;
  protected readonly faRepeat = faRepeat;
  protected readonly faTrash = faTrash;
  protected readonly faUserCircle = faUserCircle;
  protected readonly faXmark = faXmark;

  protected readonly filterForm = new FormGroup({
    search: new FormControl(''),
  });

  protected readonly users$ = filteredByFormControl(
    toObservable(this.users),
    this.filterForm.controls.search,
    (it: UserAccountWithRole, search: string) =>
      !search ||
      (it.name || '').toLowerCase().includes(search.toLowerCase()) ||
      (it.email || '').toLowerCase().includes(search.toLowerCase())
  );

  private readonly inviteUserDialog = viewChild.required<TemplateRef<unknown>>('inviteUserDialog');
  private modalRef?: DialogRef;
  protected readonly inviteForm = new FormGroup({
    email: new FormControl('', {nonNullable: true, validators: [Validators.required, Validators.email]}),
    name: new FormControl<string | undefined>(undefined, {nonNullable: true}),
    userRole: new FormControl<UserRole>('admin', {nonNullable: true, validators: [Validators.required]}),
  });
  protected inviteFormLoading = false;
  protected inviteUrl: string | null = null;

  protected readonly organization = toSignal(this.organizationService.get());

  protected readonly limit = computed(() => {
    const org = this.organization();
    return !org
      ? undefined
      : this.auth.isVendor() && this.customerOrganizationId() === undefined
        ? org.subscriptionUserAccountQuantity
        : org.subscriptionLimits.maxUsersPerCustomerOrganization;
  });

  protected readonly isProSubscription = computed(() => {
    const subscriptionType = this.organization()?.subscriptionType;
    return subscriptionType && ['trial', 'pro', 'enterprise'].includes(subscriptionType);
  });

  protected readonly editRoleUserId = signal<string | null>(null);
  protected readonly editRoleForm = new FormGroup({
    userRole: new FormControl<UserRole>('admin', {nonNullable: true, validators: [Validators.required]}),
  });
  protected editRoleFormLoading = false;

  public showInviteDialog(reset?: boolean): void {
    this.closeInviteDialog(reset);
    this.modalRef = this.overlay.showModal(this.inviteUserDialog());
  }

  protected editUserRole(user: UserAccountWithRole): void {
    if (!user.id) {
      return;
    }
    this.editRoleFormLoading = false;
    this.editRoleUserId.set(user.id);
    this.editRoleForm.reset(user);
  }

  protected async submitEditUserRoleForm(): Promise<void> {
    this.editRoleForm.markAllAsTouched();

    const userId = this.editRoleUserId();
    const userRole = this.editRoleForm.value.userRole;
    if (!userId || !userRole) {
      return;
    }

    if (this.editRoleForm.valid) {
      this.editRoleFormLoading = true;
      try {
        await firstValueFrom(this.usersService.patchUserAccount(userId, {userRole}));
        this.editRoleUserId.set(null);
        this.editRoleForm.reset();
        this.toast.success('User role has been updated');
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.editRoleFormLoading = false;
      }
    }
  }

  public async submitInviteForm(): Promise<void> {
    this.inviteForm.markAllAsTouched();
    if (this.inviteForm.valid) {
      this.inviteFormLoading = true;
      try {
        const result = await firstValueFrom(
          this.usersService.addUser({
            email: this.inviteForm.value.email!,
            name: this.inviteForm.value.name || undefined,
            userRole: this.inviteForm.value.userRole ?? 'admin',
            customerOrganizationId: this.customerOrganizationId(),
          })
        );
        this.inviteUrl = result.inviteUrl;
        if (!this.inviteUrl) {
          this.toast.success(
            `${result.user.customerOrganizationId === undefined ? 'User' : 'Customer'} has been added to the organization`
          );
          this.closeInviteDialog();
        }
        this.refresh.emit();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.inviteFormLoading = false;
      }
    }
  }

  public async uploadImage(data: UserAccountWithRole) {
    const fileId = await firstValueFrom(
      this.imageUploadService.showDialog({imageUrl: data.imageUrl, scope: 'platform'})
    );
    if (!fileId || data.imageUrl?.includes(fileId)) {
      return;
    }
    await firstValueFrom(this.usersService.patchImage(data.id!, fileId));
  }

  protected async resendInvitation(user: UserAccountWithRole) {
    try {
      const result = await firstValueFrom(this.usersService.resendInvitation(user));
      this.inviteUrl = result.inviteUrl;
      if (!this.inviteUrl) {
        this.toast.success(`Invitation has been resent to ${user.email}`);
      } else {
        this.showInviteDialog(false);
      }
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
  }

  public async deleteUser(user: UserAccountWithRole): Promise<void> {
    this.overlay
      .confirm(`This will remove ${user.name ?? user.email} from your organization. Are you sure?`)
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.usersService.delete(user)),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return NEVER;
        }),
        tap(() => this.refresh.emit())
      )
      .subscribe();
  }

  public closeInviteDialog(reset: boolean = true): void {
    this.modalRef?.close();

    if (reset) {
      this.inviteUrl = null;
      this.inviteForm.reset();
    }
  }

  public copyInviteUrl(): void {
    if (this.inviteUrl) {
      navigator.clipboard.writeText(this.inviteUrl);
      this.toast.success('Invite URL has been copied');
    }
  }
}
