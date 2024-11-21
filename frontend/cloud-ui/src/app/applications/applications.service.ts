import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Application } from '../types/application';
import { environment } from '../../env/env';
import { catchError, Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ApplicationsService {
  applicationsUrl = environment.apiBase + '/applications';

  constructor(private httpClient: HttpClient) { }

  getApplications(): Observable<Application[]> {
    // TODO some http error handling
    return this.httpClient.get<Application[]>(this.applicationsUrl)
  }
}
