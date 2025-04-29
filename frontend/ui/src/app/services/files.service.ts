import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';

@Injectable({providedIn: 'root'})
export class FilesService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/files';

  public uploadFile(file: FormData) {
    return this.httpClient.post<string>(this.baseUrl, file);
  }
}
