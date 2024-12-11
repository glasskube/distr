import {Component, inject, TemplateRef, ViewChild, ViewContainerRef} from '@angular/core';
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
import {firstValueFrom} from 'rxjs';
import {UserRole} from '../../types/user-account';

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

  private readonly users = inject(UsersService);
  public users$ = this.users.getUsers();

  private readonly overlay = inject(OverlayService);
  private readonly viewContainerRef = inject(ViewContainerRef);
  @ViewChild('inviteUserDialog') private inviteUserDialog!: TemplateRef<unknown>;
  private modalRef?: EmbeddedOverlayRef;
  public inviteForm = new FormGroup({
    email: new FormControl('', {nonNullable: true, validators: [Validators.required, Validators.email]}),
    name: new FormControl<string | undefined>(undefined, {nonNullable: true}),
    userRole: new FormControl<UserRole>('customer', {nonNullable: true, validators: [Validators.required]}),
  });

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
          userRole: this.inviteForm.value.userRole!,
        })
      );
      this.closeInviteDialog();
      this.users$ = this.users.getUsers();
    }
  }

  public closeInviteDialog(): void {
    this.modalRef?.close();
    this.inviteForm.reset();
  }
}
