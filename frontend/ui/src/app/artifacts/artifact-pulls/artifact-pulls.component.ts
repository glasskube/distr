import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, inject} from '@angular/core';
import {scan, shareReplay, startWith, Subject, switchMap, tap} from 'rxjs';
import {ArtifactPullsService} from '../../services/artifact-pulls.service';

@Component({
  templateUrl: './artifact-pulls.component.html',
  imports: [AsyncPipe, DatePipe],
})
export class ArtifactPullsComponent {
  protected hasMore = true;
  private currentOldestPull?: Date;
  private readonly fetchCount = 50;
  private readonly showMore$ = new Subject<void>();
  private readonly pulls = inject(ArtifactPullsService);

  protected readonly pulls$ = this.showMore$.pipe(
    startWith(undefined),
    switchMap(() => this.pulls.get({before: this.currentOldestPull, count: this.fetchCount})),
    tap((it) => {
      if (it.length > 0) {
        this.currentOldestPull = new Date(it[it.length - 1].createdAt);
      }
      if (it.length < this.fetchCount) {
        this.hasMore = false;
      }
    }),
    scan((all, next) => [...all, ...next]),
    shareReplay(1)
  );

  protected showMore() {
    this.showMore$.next();
  }

  protected formatRemoteAddress(addr: string): string {
    if (addr.includes(']')) {
      // IPv6
      return addr.substring(0, addr.lastIndexOf(']') + 1);
    } else if (addr.includes(':')) {
      // IPv4
      return addr.substring(0, addr.lastIndexOf(':'));
    } else {
      // fallback for undetermined format
      return addr;
    }
  }
}
