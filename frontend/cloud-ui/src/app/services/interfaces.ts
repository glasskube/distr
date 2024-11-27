import {Observable} from 'rxjs';

export interface CrudService<Request, Response = Request> {
  list(): Observable<Response[]>;
  //get(id: string): Observable<Response>;
  create(request: Request): Observable<Response>;
  update(request: Request): Observable<Response>;
}
