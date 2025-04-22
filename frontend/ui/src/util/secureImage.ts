import {inject, Pipe, PipeTransform} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {DomSanitizer, SafeUrl} from '@angular/platform-browser';
import {map, Observable, of} from 'rxjs';

@Pipe({
  name: 'secureImage'
})
export class SecureImagePipe implements PipeTransform {

  private readonly httpClient = inject(HttpClient);
  private readonly domSanitizer = inject(DomSanitizer);

  transform(url?: string): Observable<SafeUrl> {
    if (!url || !url.length) {
      return of('/distr-logo.svg');
    }
    return this.httpClient
      .get(url, {responseType: 'blob'}).pipe(
        map(val => this.domSanitizer.bypassSecurityTrustUrl(URL.createObjectURL(val)))
      );
  }
}
