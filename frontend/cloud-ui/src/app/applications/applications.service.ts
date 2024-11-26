import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, tap} from 'rxjs';
import {Application} from '../types/application';
import {DefaultReactiveList, ReactiveList} from '../services/cache';

@Injectable({
  providedIn: 'root',
})
export class ApplicationsService {
  private readonly applicationsUrl = '/api/applications';
  private readonly cache: ReactiveList<Application>;

  constructor(private readonly httpClient: HttpClient) {
    this.cache = new DefaultReactiveList(this.httpClient.get<Application[]>(this.applicationsUrl));
  }

  getApplications(): Observable<Application[]> {
    return this.cache.get();
  }

  createApplication(application: Application): Observable<Application> {
    return this.httpClient.post<Application>(this.applicationsUrl, application).pipe(tap((it) => this.cache.save(it)));
  }

  updateApplication(application: Application): Observable<Application> {
    return this.httpClient
      .put<Application>(`${this.applicationsUrl}/${application.id}`, application)
      .pipe(tap((it) => this.cache.save(it)));
  }
}
