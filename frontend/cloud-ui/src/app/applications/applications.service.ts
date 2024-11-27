import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Application, ApplicationVersion} from '../types/application';
import {combineLatest, map, Observable, scan, shareReplay, startWith, Subject, tap} from 'rxjs';

export function distinctById(applications: readonly Application[]): Application[] {
  return applications.filter((value: Application, index, self) => {
    return self.findIndex((element) => element.id === value.id) === index;
  });
}

function compareApplicationsByName(a: Application, b: Application) {
  return (a.name || '').localeCompare(b.name || '');
}

@Injectable({
  providedIn: 'root',
})
export class ApplicationsService {
  private readonly applicationsUrl = '/api/applications';

  private readonly initialApplications: Observable<Application[]>;
  private readonly savedApplications = new Subject<Application>();
  private readonly savedApplicationsAccumulated = this.savedApplications.pipe(
    scan((list: Application[], it: Application) => [it, ...list], []),
    startWith([])
  );

  private readonly applications: Observable<Application[]>;

  constructor(private readonly httpClient: HttpClient) {
    this.initialApplications = this.httpClient.get<Application[]>(this.applicationsUrl);
    this.applications = combineLatest([this.initialApplications, this.savedApplicationsAccumulated]).pipe(
      map(([initialLs, savedLs]) => distinctById([...savedLs, ...initialLs])),
      map((ls: Application[]) => ls.sort(compareApplicationsByName)),
      shareReplay(1)
    );
  }

  getApplications(): Observable<Application[]> {
    return this.applications;
  }

  createApplication(application: Application): Observable<Application> {
    return this.httpClient
      .post<Application>(this.applicationsUrl, application)
      .pipe(tap((it) => this.savedApplications.next(it)));
  }

  updateApplication(application: Application): Observable<Application> {
    return this.httpClient
      .put<Application>(`${this.applicationsUrl}/${application.id}`, application)
      .pipe(tap((it) => this.savedApplications.next(it)));
  }

  createApplicationVersion(
    application: Application,
    applicationVersion: ApplicationVersion,
    file: File
  ): Observable<ApplicationVersion> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('applicationversion', JSON.stringify(applicationVersion));
    return this.httpClient.post<ApplicationVersion>(`${this.applicationsUrl}/${application.id}/versions`, formData);
  }
}
