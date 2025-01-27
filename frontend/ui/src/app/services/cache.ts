import {combineLatest, map, Observable, scan, shareReplay, startWith, Subject} from 'rxjs';
import {compareBy, distinctBy, Predicate} from '../../util/arrays';
import {BaseModel, Named} from '@glasskube/distr-sdk';

export abstract class ReactiveList<T> {
  protected abstract readonly identify: Predicate<T, unknown>;
  protected abstract readonly sortAttr: Predicate<T, string>;

  private readonly saved$ = new Subject<T>();
  private readonly savedAcc$: Observable<T[]> = this.saved$.pipe(
    scan((list: T[], it: T) => [it, ...list], []),
    startWith([])
  );

  private readonly removed$ = new Subject<T>();
  private readonly removedIdsAcc$: Observable<Set<unknown>> = this.removed$.pipe(
    map((it) => this.identify(it)),
    scan((set: Set<unknown>, it: unknown) => set.add(it), new Set()),
    startWith(new Set())
  );

  private readonly actual$: Observable<T[]>;

  constructor(private readonly initial$: Observable<T[]>) {
    this.actual$ = combineLatest([this.initial$, this.savedAcc$, this.removedIdsAcc$]).pipe(
      map(([initialLs, savedLs, removedIds]) =>
        distinctBy(this.identify)([...savedLs, ...initialLs]).filter((it) => !removedIds.has(this.identify(it)))
      ),
      map((ls: T[]) => ls.sort(compareBy(this.sortAttr))),
      shareReplay(1)
    );
  }

  public save(arg: T) {
    this.saved$.next(arg);
  }

  public remove(arg: T) {
    this.removed$.next(arg);
  }

  public get(): Observable<T[]> {
    return this.actual$;
  }
}

export class DefaultReactiveList<T extends Named & BaseModel> extends ReactiveList<T> {
  protected override identify = (it: T) => it.id;
  protected override sortAttr = (it: T) => it.name!;
}
