import {Injectable} from '@angular/core';
import {License} from '../types/license';
import {from, Observable} from 'rxjs';

@Injectable({providedIn: 'root'})
export class LicenseService {
  public getLicensesForApplication(applicationId: string): Observable<License[]> {
    return from([
      [
        {
          id: 'ced3992d-6717-498e-8726-6845ec3069b4',
          name: 'test',
          applicationId,
          versions: [
            // {id: 'bf58c94e-0c5b-476a-a647-09464d85bcae', name: 'v4.2.0'}
          ],
        },
      ],
    ]);
  }
}
