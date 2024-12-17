import {combineLatest, map, Observable, startWith} from 'rxjs';
import {FormControl} from '@angular/forms';

export function filteredByFormControl<T>(
  dataSource: Observable<T[]>,
  formControl: FormControl,
  matchFn: (item: T, search: string) => boolean
): Observable<T[]> {
  return combineLatest([dataSource, formControl.valueChanges.pipe(startWith(''))]).pipe(
    map(([items, search]) => {
      return items.filter((it) => matchFn(it, search));
    })
  );
}
