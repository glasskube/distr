import {Component} from '@angular/core';
import {UsersComponent} from '../users.component';

@Component({
  template: ` <section class="bg-gray-50 dark:bg-gray-900 p-3 sm:p-5 antialiased sm:ml-64">
    <div class="mx-auto max-w-screen-2xl px-4 lg:px-12">
      <div class="bg-white dark:bg-gray-800 relative shadow-md sm:rounded-lg overflow-hidden">
        <app-users />
      </div>
    </div>
  </section>`,
  imports: [UsersComponent],
})
export class VendorUsersComponent {}
