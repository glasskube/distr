import {HttpClient} from '@angular/common/http';
import {inject, Pipe, PipeTransform} from '@angular/core';
import {SafeUrl} from '@angular/platform-browser';
import {map, Observable} from 'rxjs';

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

@Pipe({name: 'secureImage'})
export class SecureImagePipe implements PipeTransform {
  private readonly httpClient = inject(HttpClient);

  transform(urlOrUuid: string): Observable<SafeUrl> {
    if (uuidPattern.test(urlOrUuid)) {
      urlOrUuid = '/api/v1/files/' + urlOrUuid;
    }
    return this.httpClient.get(urlOrUuid, {responseType: 'blob'}).pipe(map((val) => URL.createObjectURL(val)));
  }
}
