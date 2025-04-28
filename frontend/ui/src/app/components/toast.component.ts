import {animate, keyframes, state, style, transition, trigger} from '@angular/animations';
import {Component} from '@angular/core';
import {Toast} from 'ngx-toastr';
import {faCheck, faCircleExclamation} from '@fortawesome/free-solid-svg-icons';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';

@Component({
  selector: '[toast-component]',
  styles: [
    `
      :host {
        pointer-events: all;
        background-color: transparent;
        background-image: none;
      }
    `,
  ],
  template: `
    <div
      [class.border-red-300]="options.payload === 'error'"
      [class.dark:border-red-800]="options.payload === 'error'"
      [class.border-green-300]="options.payload === 'success'"
      [class.dark:border-green-800]="options.payload === 'success'"
      class="flex items-center w-full max-w-xs p-4 mb-4 text-gray-500 bg-white rounded-lg shadow-sm dark:text-gray-400 dark:bg-gray-800 border border-gray-200 dark:border-gray-600"
      role="alert">
      @switch (options.payload) {
        @case ('error') {
          <fa-icon
            [icon]="faCircleExclamation"
            size="lg"
            class="inline-flex items-center justify-center flex-shrink-0 w-8 h-8 rounded-lg text-red-500 dark:bg-red-800 bg-red-100 dark:text-red-200">
          </fa-icon>
        }
        @case ('success') {
          <fa-icon
            [icon]="faCheck"
            size="lg"
            class="inline-flex items-center justify-center flex-shrink-0 w-8 h-8 rounded-lg text-green-500 dark:text-green-800">
          </fa-icon>
        }
      }
      <div class="ms-3 text-sm font-normal">
        {{ title }}
        @if (message) {
          :{{ message }}
        }
      </div>
      <button
        type="button"
        (click)="remove()"
        class="ms-auto -mx-1.5 -my-1.5 bg-white text-gray-400 hover:text-gray-900 rounded-lg focus:ring-2 focus:ring-gray-300 p-1.5 hover:bg-gray-100 inline-flex items-center justify-center h-8 w-8 dark:text-gray-500 dark:hover:text-white dark:bg-gray-800 dark:hover:bg-gray-700"
        data-dismiss-target="#toast-danger"
        aria-label="Close">
        <span class="sr-only">Close</span>
        <svg class="w-3 h-3" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 14">
          <path
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6" />
        </svg>
      </button>
    </div>
  `,
  animations: [
    trigger('flyInOut', [
      state(
        'inactive',
        style({
          opacity: 0,
        })
      ),
      transition(
        'inactive => active',
        animate(
          '400ms ease-out',
          keyframes([
            style({
              transform: 'translate3d(100%, 0, 0) skewX(-30deg)',
              opacity: 0,
            }),
            style({
              transform: 'skewX(20deg)',
              opacity: 1,
            }),
            style({
              transform: 'skewX(-5deg)',
              opacity: 1,
            }),
            style({
              transform: 'none',
              opacity: 1,
            }),
          ])
        )
      ),
      transition(
        'active => removed',
        animate(
          '400ms ease-out',
          keyframes([
            style({
              opacity: 1,
            }),
            style({
              transform: 'translate3d(100%, 0, 0) skewX(30deg)',
              opacity: 0,
            }),
          ])
        )
      ),
    ]),
  ],
  preserveWhitespaces: false,
  imports: [FaIconComponent],
})
export class ToastComponent extends Toast {
  protected readonly faCheck = faCheck;
  protected readonly faCircleExclamation = faCircleExclamation;
}
