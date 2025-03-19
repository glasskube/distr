import {Component, computed, inject, input, signal} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faDownload, faEllipsis} from '@fortawesome/free-solid-svg-icons';
import {HasDownloads} from '../services/artifacts.service';
import {UsersService} from '../services/users.service';
import {toObservable} from '@angular/core/rxjs-interop';
import {switchMap, zip} from 'rxjs';
import {AsyncPipe} from '@angular/common';

@Component({
  selector: 'app-artifacts-download-count',
  template: `
    <div class="inline-flex items-center text-sm text-gray-500 truncate dark:text-gray-400">
      <fa-icon class="me-1" [icon]="faDownload"></fa-icon>
      {{ source().downloadsTotal }}
    </div>
  `,
  imports: [FaIconComponent],
})
export class ArtifactsDownloadCountComponent {
  public readonly source = input.required<HasDownloads>();

  protected readonly faDownload = faDownload;
}

@Component({
  selector: 'app-artifacts-downloaded-by',
  template: `
    <div class="flex -space-x-3 hover:-space-x-1 rtl:space-x-reverse">
      @for (user of downloadedBy$ | async; track user?.user?.id) {
        <img
          class="size-8 border-2 border-white rounded-full dark:border-gray-800 transition-all duration-100 ease-in-out"
          [src]="user?.gravatar"
          [title]="user?.user?.name ?? user?.user?.email" />
      }
      @if ((source().downloadedByCount ?? 0) - (source().downloadedByUsers ?? []).length; as count) {
        @if (count > 0) {
          <div
            class="flex items-center justify-center size-8 text-xs font-medium text-white bg-gray-500 dark:bg-gray-700 border-2 border-white rounded-full dark:border-gray-800">
            +{{ count }}
          </div>
        }
      }
    </div>
  `,
  imports: [AsyncPipe],
})
export class ArtifactsDownloadedByComponent {
  public readonly source = input.required<HasDownloads>();
  private readonly usersService = inject(UsersService);
  public readonly downloadedBy$ = toObservable(this.source).pipe(
    switchMap((dl) => {
      const gravatarObservables = (dl.downloadedByUsers ?? []).map((id) =>
        this.usersService.getUserWithGravatarUrl(id)
      );
      return zip(...gravatarObservables);
    })
  );
}

@Component({
  selector: 'app-artifacts-hash',
  template: `
    <span class="font-mono">{{ hashForDisplay() }}</span>
    @if (expandable()) {
      <button
        type="button"
        class="inline-flex items-center justify-center h-3.5 ms-1 px-1 rounded-sm bg-gray-200 hover:bg-gray-100 dark:bg-gray-700 dark:hover:bg-gray-600"
        (click)="showFull.set(!showFull())">
        <fa-icon [icon]="faEllipsis"></fa-icon>
      </button>
    }
  `,
  imports: [FaIconComponent],
})
export class ArtifactsHashComponent {
  public readonly hash = input.required<string>();
  public readonly expandable = input<boolean>(true);
  protected readonly showFull = signal(false);
  protected readonly hashForDisplay = computed(() => (this.showFull() ? this.hash() : this.hash().substring(0, 17)));

  protected readonly faEllipsis = faEllipsis;
}
