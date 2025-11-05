import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, computed, inject, input, OnDestroy, Signal, TemplateRef, ViewChild} from '@angular/core';
import {takeUntilDestroyed, toObservable, toSignal} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faBox,
  faCircleExclamation,
  faClipboard,
  faMagnifyingGlass,
  faPlus,
  faRepeat,
  faTrash,
  faUserCircle,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {UserAccountWithRole, UserRole} from '@glasskube/distr-sdk';
import {
  catchError,
  combineLatest,
  filter,
  firstValueFrom,
  map,
  NEVER,
  Observable,
  startWith,
  Subject,
  switchMap,
  tap,
} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {filteredByFormControl} from '../../../util/filter';
import {SecureImagePipe} from '../../../util/secureImage';
import {modalFlyInOut} from '../../animations/modal';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {UsersService} from '../../services/users.service';
import {UuidComponent} from '../uuid';

@Component({
  selector: 'app-users',
  imports: [
    FaIconComponent,
    AsyncPipe,
    DatePipe,
    ReactiveFormsModule,
    RequireRoleDirective,
    AutotrimDirective,
    UuidComponent,
    SecureImagePipe,
  ],
  templateUrl: './users.component.html',
  animations: [modalFlyInOut],
})
export class UsersComponent implements OnDestroy {
  private readonly toast = inject(ToastService);
  private readonly users = inject(UsersService);
  private readonly overlay = inject(OverlayService);
  readonly featureFlags = inject(FeatureFlagService);

  public readonly faMagnifyingGlass = faMagnifyingGlass;
  public readonly faPlus = faPlus;
  protected readonly faRepeat = faRepeat;
  public readonly faXmark = faXmark;
  protected readonly faTrash = faTrash;
  protected readonly faClipboard = faClipboard;

  public readonly customerOrganizationId = input<string>();
  public readonly userRole: Signal<UserRole> = computed(() =>
    this.customerOrganizationId() === undefined ? 'vendor' : 'customer'
  );
  public readonly users$: Observable<UserAccountWithRole[]>;
  private readonly refresh$ = new Subject<void>();

  @ViewChild('inviteUserDialog') private inviteUserDialog!: TemplateRef<unknown>;
  private modalRef?: DialogRef;
  public inviteForm = new FormGroup({
    email: new FormControl('', {nonNullable: true, validators: [Validators.required, Validators.email]}),
    name: new FormControl<string | undefined>(undefined, {nonNullable: true}),
  });
  inviteFormLoading = false;
  protected inviteUrl: string | null = null;

  filterForm = new FormGroup({
    search: new FormControl(''),
  });

  constructor() {
    const usersWithRefresh = this.refresh$.pipe(
      startWith(undefined),
      switchMap(() => this.users.getUsers())
    );
    const shownUserAccounts = combineLatest([toObservable(this.customerOrganizationId), usersWithRefresh]).pipe(
      map(([customerOrganizationId, users]) =>
        users.filter((it) => it.customerOrganizationId === customerOrganizationId)
      )
    );
    this.users$ = filteredByFormControl(
      shownUserAccounts,
      this.filterForm.controls.search,
      (it: UserAccountWithRole, search: string) =>
        !search ||
        (it.name || '').toLowerCase().includes(search.toLowerCase()) ||
        (it.email || '').toLowerCase().includes(search.toLowerCase())
    ).pipe(takeUntilDestroyed());
  }

  ngOnDestroy() {
    this.refresh$.complete();
  }

  public showInviteDialog(reset?: boolean): void {
    this.closeInviteDialog(reset);
    this.modalRef = this.overlay.showModal(this.inviteUserDialog);
  }

  public async submitInviteForm(): Promise<void> {
    this.inviteForm.markAllAsTouched();
    if (this.inviteForm.valid) {
      this.inviteFormLoading = true;
      try {
        const result = await firstValueFrom(
          this.users.addUser({
            email: this.inviteForm.value.email!,
            name: this.inviteForm.value.name || undefined,
            userRole: this.userRole(),
            customerOrganizationId: this.customerOrganizationId(),
          })
        );
        this.inviteUrl = result.inviteUrl;
        if (!this.inviteUrl) {
          this.toast.success(
            `${this.userRole() === 'vendor' ? 'User' : 'Customer'} has been added to the organization`
          );
          this.closeInviteDialog();
        }
        this.refresh$.next();
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
    const fileId = await firstValueFrom(this.overlay.uploadImage({imageUrl: data.imageUrl, scope: 'platform'}));
    if (!fileId || data.imageUrl?.includes(fileId)) {
      return;
    }
    await firstValueFrom(this.users.patchImage(data.id!, fileId));
  }

  protected async resendInvitation(user: UserAccountWithRole) {
    try {
      const result = await firstValueFrom(this.users.resendInvitation(user));
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
        switchMap(() => this.users.delete(user)),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return NEVER;
        }),
        tap(() => this.refresh$.next())
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

  protected readonly faCircleExclamation = faCircleExclamation;
  protected readonly faUserCircle = faUserCircle;
  protected readonly faBox = faBox;
}
