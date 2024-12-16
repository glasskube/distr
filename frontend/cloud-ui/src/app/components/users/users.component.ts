import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, computed, inject, Signal, TemplateRef, ViewChild} from '@angular/core';
import {toObservable, toSignal} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faMagnifyingGlass, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, firstValueFrom, map, Observable, startWith, Subject, switchMap} from 'rxjs';
import {modalFlyInOut} from '../../animations/modal';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {EmbeddedOverlayRef, OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {UsersService} from '../../services/users.service';
import {UserAccount, UserAccountWithRole, UserRole} from '../../types/user-account';

@Component({
  selector: 'app-users',
  imports: [FaIconComponent, AsyncPipe, DatePipe, ReactiveFormsModule, RequireRoleDirective],
  templateUrl: './users.component.html',
  animations: [modalFlyInOut],
})
export class UsersComponent {
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

  @ViewChild('inviteUserDialog') private inviteUserDialog!: TemplateRef<unknown>;
  private modalRef?: EmbeddedOverlayRef;
  public inviteForm = new FormGroup({
    email: new FormControl('', {nonNullable: true, validators: [Validators.required, Validators.email]}),
    name: new FormControl<string | undefined>(undefined, {nonNullable: true}),
  });

  constructor() {
    const data = toSignal(inject(ActivatedRoute).data);
    this.userRole = computed(() => data()?.['userRole'] ?? null);
    const usersWithRefresh = this.refresh$.pipe(
      startWith(undefined),
      switchMap(() => this.users.getUsers())
    );
    this.users$ = combineLatest([toObservable(this.userRole), usersWithRefresh]).pipe(
      map(([userRole, users]) => users.filter((it) => userRole !== null && it.userRole === userRole))
    );
  }

  public showInviteDialog(): void {
    this.closeInviteDialog();
    this.modalRef = this.overlay.showModal(this.inviteUserDialog);
  }

  public async submitInviteForm(): Promise<void> {
    this.inviteForm.markAllAsTouched();
    if (this.inviteForm.valid) {
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
    }
  }

  public async deleteUser(user: UserAccount): Promise<void> {
    if (confirm(`Really delete ${user.name ?? user.email}? This will also delete all deployments they manage!`)) {
      await firstValueFrom(this.users.delete(user));
      this.refresh$.next();
    }
  }

  public closeInviteDialog(): void {
    this.modalRef?.close();
    this.inviteForm.reset();
  }
}
