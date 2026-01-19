import {HttpClient} from '@angular/common/http';
import {inject, Pipe, PipeTransform} from '@angular/core';
import {SafeUrl} from '@angular/platform-browser';
import {map, Observable} from 'rxjs';

@Pipe({name: 'secureImage'})
export class SecureImagePipe implements PipeTransform {
  private readonly httpClient = inject(HttpClient);

  transform(url: string): Observable<SafeUrl> {
    return this.httpClient.get(url, {responseType: 'blob'}).pipe(map((val) => URL.createObjectURL(val)));
  }
}
