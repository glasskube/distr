import {Component, inject} from '@angular/core';
import {startWith, Subject, switchMap} from 'rxjs';
import {UsersService} from '../../../services/users.service';
import {UsersComponent} from '../users.component';
import {toSignal} from '@angular/core/rxjs-interop';
import {AuthService} from '../../../services/auth.service';

@Component({
  template: `<section class="bg-gray-50 dark:bg-gray-900 p-3 sm:p-5 antialiased sm:ml-64">
    <div class="mx-auto max-w-screen-2xl px-4 lg:px-12">
      <div class="bg-white dark:bg-gray-800 relative shadow-md sm:rounded-lg overflow-hidden">
        <app-users (refresh)="refresh$.next()" [users]="users() ?? []" [userRole]="userRole" />
      </div>
    </div>
  </section>`,
  imports: [UsersComponent],
})
export class VendorUsersComponent {
  private readonly usersService = inject(UsersService);
  private readonly auth = inject(AuthService);
  protected readonly refresh$ = new Subject<void>();
  protected readonly users = toSignal(
    this.refresh$.pipe(
      startWith(undefined),
      switchMap(() => this.usersService.getUsers())
    )
  );
  protected readonly userRole = this.auth.getClaims()!.role;
}
