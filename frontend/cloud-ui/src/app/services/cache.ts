import {Subject, Observable, scan, startWith, combineLatest, map, shareReplay} from 'rxjs';
import {BaseModel, Named} from '../types/base';
import {compareByName, distinctById} from '../../util/arrays';

export abstract class ReactiveList<T> {
  protected abstract readonly distinctFn: (arg: T[]) => T[];
  protected abstract readonly compareFn: (a: T, b: T) => number;
  private readonly saved$ = new Subject<T>();
  private readonly savedAcc$: Observable<T[]> = this.saved$.pipe(
    scan((list: T[], it: T) => [it, ...list], []),
    startWith([])
  );
  private readonly actual$: Observable<T[]>;
  constructor(private readonly initial$: Observable<T[]>) {
    this.actual$ = combineLatest([this.initial$, this.savedAcc$]).pipe(
      map(([initialLs, savedLs]) => this.distinctFn([...savedLs, ...initialLs])),
      map((ls: T[]) => ls.sort(this.compareFn)),
      shareReplay(1)
    );
  }

  public save(arg: T) {
    this.saved$.next(arg);
  }

  public get(): Observable<T[]> {
    return this.actual$;
  }
}

export class DefaultReactiveList<T extends Named & BaseModel> extends ReactiveList<T> {
  protected override compareFn = compareByName;
  protected override distinctFn = distinctById;
}
