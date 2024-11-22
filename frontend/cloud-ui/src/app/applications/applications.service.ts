import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Application } from '../types/application';
import { combineLatest, map, Observable, scan, shareReplay, startWith, Subject, tap } from 'rxjs';

const accumulate = scan((list: Application[], it: Application) => [it, ...list], [] as Application[]);

type ElementSelectFn<T> = (element: T) => unknown
type ElementSelector<T> = keyof T | ElementSelectFn<T>

function selectFn<T>(selector: ElementSelector<T>): ElementSelectFn<T> {
  return selector instanceof Function ? selector : element => element[selector];
}

export function distinctBy<T>(array: readonly T[], selector: ElementSelector<T>): T[] {
  const select = selectFn(selector);
  return array.filter((value, index, self) => {
    const transformedValue = select(value);
    return self.findIndex(element => select(element) === transformedValue) === index;
  });
}

@Injectable({
  providedIn: 'root',
})
export class ApplicationsService {
  private readonly applicationsUrl = '/api/applications';

  private initialApplications: Observable<Application[]>;
  private readonly savedApplications = new Subject<Application>();
  private readonly savedApplicationsAccumulated = this.savedApplications.pipe(
    scan((list: Application[], it: Application) => [it, ...list], []), startWith([]))

  private readonly applications: Observable<Application[]>;

  constructor(private readonly httpClient: HttpClient) {
    this.initialApplications = this.httpClient.get<Application[]>(this.applicationsUrl);
    this.applications = combineLatest([this.initialApplications, this.savedApplicationsAccumulated]).pipe(
      map(([initialLs, savedLs]) => distinctBy([...savedLs, ...initialLs], 'id')),
      shareReplay(1)
    )
  }

  getApplications(): Observable<Application[]> {
    return this.applications;
  }

  createApplication(application: Application): Observable<Application> {
    return this.httpClient.post<Application>(this.applicationsUrl, application)
      .pipe(tap(it => this.savedApplications.next(it)));
  }

  updateApplication(application: Application): Observable<Application> {
    return this.httpClient.put<Application>(`${this.applicationsUrl}/${application.id}`, application)
      .pipe(tap(it => this.savedApplications.next(it)));
  }
}
