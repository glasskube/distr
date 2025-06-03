import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';

export type FileScope = 'platform' | 'organization';

@Injectable({providedIn: 'root'})
export class FilesService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/files';

  public uploadFile(file: FormData, scope?: FileScope) {
    return this.httpClient.post<string>(`${this.baseUrl}${!!scope ? '?scope=' + scope : ''}`, file);
  }
}
