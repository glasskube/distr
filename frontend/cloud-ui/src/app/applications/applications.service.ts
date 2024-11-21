import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Application } from '../types/application';
import { environment } from '../../env/env';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ApplicationsService {
  applicationsUrl = environment.apiBase + '/applications';

  constructor(private httpClient: HttpClient) { }

  getApplications(): Observable<Application[]> {
    return this.httpClient.get<Application[]>(this.applicationsUrl)
  }
}
