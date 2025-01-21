import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, computed, inject, OnDestroy, Signal, TemplateRef, ViewChild} from '@angular/core';
import {toObservable, toSignal} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faMagnifyingGlass, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {
  combineLatest,
  filter,
  firstValueFrom,
  map,
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
import {UserAccount, UserAccountWithRole, UserRole} from '../../types/user-account';
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
  ],
  templateUrl: './users.component.html',
  animations: [modalFlyInOut],
})
export class UsersComponent implements OnDestroy {
  private readonly toast = inject(ToastService);
  private readonly users = inject(UsersService);
  private readonly overlay = inject(OverlayService);

  public readonly faMagnifyingGlass = faMagnifyingGlass;
  public readonly faPlus = faPlus;
  public readonly faXmark = faXmark;
  protected readonly faTrash = faTrash;

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
        await firstValueFrom(
          this.users.addUser({
            email: this.inviteForm.value.email!,
            name: this.inviteForm.value.name || undefined,
            userRole: this.userRole(),
          })
        );
        this.closeInviteDialog();
        switch (this.userRole()) {
          case 'customer':
            this.toast.success('Customer has been invited to the organization');
            break;
          case 'vendor':
            this.toast.success('User has been invited to the organization');
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

  public async deleteUser(user: UserAccount): Promise<void> {
    this.overlay
      .confirm(`Really delete ${user.name ?? user.email}? This will also delete all deployments they manage!`)
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.users.delete(user)),
        tap(() => this.refresh$.next())
      )
      .subscribe();
  }

  public closeInviteDialog(): void {
    this.modalRef?.close();
    this.inviteForm.reset();
  }
}
