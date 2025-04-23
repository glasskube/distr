import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, computed, inject, OnDestroy, Signal, TemplateRef, ViewChild} from '@angular/core';
import {toObservable, toSignal} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faBox,
  faCircleExclamation,
  faClipboard,
  faMagnifyingGlass,
  faPlus,
  faTrash, faUserCircle,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
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
  takeUntil,
  tap,
} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {filteredByFormControl} from '../../../util/filter';
import {modalFlyInOut} from '../../animations/modal';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {UsersService} from '../../services/users.service';
import {UuidComponent} from '../uuid';
import {UserAccountWithRole, UserRole} from '@glasskube/distr-sdk';
import {HttpErrorResponse} from '@angular/common/http';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {SecureImagePipe} from '../../../util/secureImage';

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
  public readonly faXmark = faXmark;
  protected readonly faTrash = faTrash;
  protected readonly faClipboard = faClipboard;

  public readonly userRole: Signal<UserRole>;
  public readonly users$: Observable<UserAccountWithRole[]>;
  private readonly refresh$ = new Subject<void>();
  private readonly destroyed$ = new Subject<void>();

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
    const data = toSignal(inject(ActivatedRoute).data);
    this.userRole = computed(() => data()?.['userRole'] ?? null);
    const usersWithRefresh = this.refresh$.pipe(
      startWith(undefined),
      switchMap(() => this.users.getUsers())
    );
    const shownUserAccounts = combineLatest([toObservable(this.userRole), usersWithRefresh]).pipe(
      map(([userRole, users]) => users.filter((it) => userRole !== null && it.userRole === userRole))
    );
    this.users$ = filteredByFormControl(
      shownUserAccounts,
      this.filterForm.controls.search,
      (it: UserAccountWithRole, search: string) =>
        !search ||
        (it.name || '').toLowerCase().includes(search.toLowerCase()) ||
        (it.email || '').toLowerCase().includes(search.toLowerCase())
    ).pipe(takeUntil(this.destroyed$));
  }

  ngOnDestroy() {
    this.refresh$.complete();
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  public showInviteDialog(): void {
    this.closeInviteDialog();
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
          })
        );
        this.inviteUrl = result.inviteUrl;
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
    const fileId = await firstValueFrom(this.overlay.uploadImage({imageUrl: data.imageUrl}));
    if (!fileId || data.imageUrl?.includes(fileId)) {
      return;
    }
    await firstValueFrom(this.users.patchImage(data.id!!, fileId));
  }

  public async deleteUser(user: UserAccountWithRole): Promise<void> {
    this.overlay
      .confirm(`Really delete ${user.name ?? user.email}?`)
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.users.delete(user)),
        catchError((e) => {
          if (e instanceof HttpErrorResponse && e.status === 400) {
            this.toast.error(
              `User ${user.name ?? user.email} cannot be deleted.
              Please ensure there are no deployments managed by this user and try again.`
            );
          }
          return NEVER;
        }),
        tap(() => this.refresh$.next())
      )
      .subscribe();
  }

  public closeInviteDialog(): void {
    this.inviteUrl = null;
    this.modalRef?.close();
    this.inviteForm.reset();
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
