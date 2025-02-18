import {Component, inject, Input} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faClipboardCheck} from '@fortawesome/free-solid-svg-icons';
import {faClipboard} from '@fortawesome/free-regular-svg-icons';
import {ToastService} from '../services/toast.service';

@Component({
  selector: 'app-uuid',
  template: `
    <button
      (click)="copyUuid()"
      [title]="uuid"
      type="button"
      [class.text-xs]="small"
      [class.py-0.5]="small"
      [class.py-2]="!small"
      [class.rounded]="small"
      [class.rounded-lg]="!small"
      class="text-gray-900 dark:text-gray-400 m-0.5 hover:bg-gray-100 dark:bg-gray-800 dark:border-gray-600 dark:hover:bg-gray-700 px-2.5 inline-flex items-center justify-center bg-white border-gray-200 border">
      <span class="inline-flex items-center">
        <code>{{ shortUuid }}</code>
        @if (!copied) {
          <fa-icon [icon]="faClipboard" class="ml-2" />
        } @else {
          <fa-icon [icon]="faClipboardCheck" class="ml-2" />
        }
      </span>
    </button>
  `,
  imports: [FaIconComponent],
})
export class UuidComponent {
  @Input({required: true})
  uuid!: string;

  @Input()
  small = false;

  protected copied = false;

  private readonly toast = inject(ToastService);

  protected get shortUuid(): string {
    return this.uuid.slice(0, 8);
  }

  protected async copyUuid() {
    await navigator.clipboard.writeText(this.uuid);
    this.toast.success('ID copied to clipboard');
    this.copied = true;
    setTimeout(() => (this.copied = false), 2000);
  }

  protected readonly faClipboard = faClipboard;
  protected readonly faClipboardCheck = faClipboardCheck;
}
