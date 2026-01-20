import {HttpClient} from '@angular/common/http';
import {inject, OnDestroy, Pipe, PipeTransform} from '@angular/core';
import {SafeUrl} from '@angular/platform-browser';
import {map, Observable, tap} from 'rxjs';

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

@Pipe({name: 'secureImage'})
export class SecureImagePipe implements PipeTransform, OnDestroy {
  private readonly httpClient = inject(HttpClient);

  private previousObjectUrl?: string;

  public ngOnDestroy(): void {
    this.revokePreviousObjectUrl();
  }

  public transform(urlOrUuid: string): Observable<SafeUrl> {
    if (uuidPattern.test(urlOrUuid)) {
      urlOrUuid = '/api/v1/files/' + urlOrUuid;
    }
    return this.httpClient.get(urlOrUuid, {responseType: 'blob'}).pipe(
      map((data) => URL.createObjectURL(data)),
      tap((objectUrl) => {
        this.revokePreviousObjectUrl();
        this.previousObjectUrl = objectUrl;
      })
    );
  }

  private revokePreviousObjectUrl(): void {
    if (this.previousObjectUrl) {
      URL.revokeObjectURL(this.previousObjectUrl);
    }
  }
}
