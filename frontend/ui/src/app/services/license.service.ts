import {Injectable} from '@angular/core';
import {License} from '../types/license';
import {from, Observable} from 'rxjs';

@Injectable({providedIn: 'root'})
export class LicenseService {
  public getLicensesForApplication(applicationId: string): Observable<License[]> {
    return from([
      [
        {
          id: 'a',
          name: 'L1',
          versions: [
            {id: 'a', name: 'V1'},
            {id: 'c', name: 'V3'},
          ],
        },
        {id: 'b', name: 'L2', versions: [{id: 'b', name: 'V2'}]},
      ],
    ]);
  }
}
