import {concat, map, Observable, scan, shareReplay, Subject} from 'rxjs';
import {compareBy, distinctBy, Predicate} from '../../util/arrays';
import {BaseModel, Named} from '@glasskube/distr-sdk';

interface ReactiveListEvent<T> {
  objects?: T[]; // used for type 'reset'
  object?: T; // used when type is 'save' or 'remove'
  type: ReactiveListEventType;
}

type ReactiveListEventType = 'save' | 'remove' | 'reset';

export abstract class ReactiveList<T> {
  protected abstract readonly identify: Predicate<T, unknown>;
  protected abstract readonly sortAttr: Predicate<T, string>;

  private readonly events$ = new Subject<ReactiveListEvent<T>>();
  private readonly state$: Observable<T[]>;

  constructor(private readonly initial$: Observable<T[]>) {
    // TODO potential race condition: initial load takes too long and other stuff is being added locally
    // could be prevented locally by requiring the first event to always be of reset type and otherwise not pushing the event?
    this.state$ = concat(
      this.initial$.pipe(map((items) => ({type: 'reset', objects: items}) as ReactiveListEvent<T>)),
      this.events$
    ).pipe(
      scan((state: T[], event: ReactiveListEvent<T>) => {
        if (event.type === 'reset' && event.objects) {
          return event.objects;
        } else if (event.type === 'save' && event.object) {
          return distinctBy(this.identify)([event.object, ...state]);
        } else if (event.type === 'remove' && event.object) {
          return state.filter((item) => {
            return this.identify(item) !== this.identify(event.object!);
          });
        } else {
          console.warn('unsupported/invalid event: ' + event);
          return state;
        }
      }, []),
      map((ls: T[]) => ls.sort(compareBy(this.sortAttr))),
      shareReplay(1)
    );
  }

  public save(arg: T) {
    this.events$.next({
      type: 'save',
      object: arg,
    });
  }

  public remove(arg: T) {
    this.events$.next({
      type: 'remove',
      object: arg,
    });
  }

  public reset(arg: T[]) {
    this.events$.next({
      type: 'reset',
      objects: arg,
    });
  }

  public get(): Observable<T[]> {
    return this.state$;
  }
}

export class DefaultReactiveList<T extends Named & BaseModel> extends ReactiveList<T> {
  protected override identify = (it: T) => it.id;
  protected override sortAttr = (it: T) => it.name!;
}
