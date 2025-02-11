import {inject, Injectable} from '@angular/core';
import {map} from 'rxjs';
import {OrganizationService} from './organization.service';

@Injectable({
  providedIn: 'root',
})
export class FeatureFlagService {
  private readonly organizationService = inject(OrganizationService);
  public isLicensingEnabled$ = this.organizationService
    .get()
    .pipe(map((o) => (o.features ?? []).includes('licensing')));
}
