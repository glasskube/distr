import {Component, computed, inject, Signal, TemplateRef, ViewChild, ViewContainerRef} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faMagnifyingGlass,
  faPlus,
  faCaretDown,
  faPen,
  faTrash,
  faXmark,
  faBoxArchive,
} from '@fortawesome/free-solid-svg-icons';
import {UsersService} from '../../services/users.service';
import {AsyncPipe, DatePipe, JsonPipe} from '@angular/common';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {EmbeddedOverlayRef, OverlayService} from '../../services/overlay.service';
import {modalFlyInOut} from '../../animations/modal';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {combineLatest, firstValueFrom, map, Observable, startWith, Subject, switchMap, tap} from 'rxjs';
import {UserAccount, UserAccountWithRole, UserRole} from '../../types/user-account';
import {ActivatedRoute} from '@angular/router';
import {toObservable, toSignal} from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-users',
  imports: [FaIconComponent, AsyncPipe, DatePipe, RequireRoleDirective, ReactiveFormsModule],
  templateUrl: './users.component.html',
  animations: [modalFlyInOut],
})
export class UsersComponent {
  public readonly magnifyingGlassIcon = faMagnifyingGlass;
  public readonly plusIcon = faPlus;
  public readonly caretDownIcon = faCaretDown;
  public readonly penIcon = faPen;
  public readonly trashIcon = faTrash;
  public readonly xmarkIcon = faXmark;
  public readonly releaseIcon = faBoxArchive;

  public readonly userRole: Signal<UserRole>;
  private readonly users = inject(UsersService);
  public users$: Observable<UserAccountWithRole[]>;
  private readonly refresh$ = new Subject<void>();

  private readonly overlay = inject(OverlayService);
  private readonly viewContainerRef = inject(ViewContainerRef);
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
    this.modalRef = this.overlay.showModal(this.inviteUserDialog, this.viewContainerRef);
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
      this.refresh$.next();
    }
  }

  public closeInviteDialog(): void {
    this.modalRef?.close();
    this.inviteForm.reset();
  }
}
