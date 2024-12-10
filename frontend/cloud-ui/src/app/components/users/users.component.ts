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
import {AsyncPipe, DatePipe} from '@angular/common';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {EmbeddedOverlayRef, OverlayService} from '../../services/overlay.service';

@Component({
  selector: 'app-users',
  imports: [FaIconComponent, AsyncPipe, DatePipe, RequireRoleDirective],
  templateUrl: './users.component.html',
})
export class UsersComponent {
  magnifyingGlassIcon = faMagnifyingGlass;
  plusIcon = faPlus;
  caretDownIcon = faCaretDown;
  penIcon = faPen;
  trashIcon = faTrash;
  xmarkIcon = faXmark;
  releaseIcon = faBoxArchive;
  private readonly users = inject(UsersService);
  private readonly overlay = inject(OverlayService);
  private readonly viewContainerRef = inject(ViewContainerRef);
  @ViewChild('inviteUserDialog') inviteUserDialog!: TemplateRef<unknown>;
  public readonly users$ = this.users.getUsers();
  private modal?: EmbeddedOverlayRef;

  public showInviteDialog(): void {
    this.modal?.close();
    this.modal = this.overlay.showModal(this.inviteUserDialog, this.viewContainerRef);
  }
}
