import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, inject, OnDestroy, OnInit, TemplateRef} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faDownload, faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {
  catchError,
  combineLatest,
  debounceTime,
  EMPTY,
  filter,
  first,
  firstValueFrom,
  map,
  Observable,
  shareReplay,
  startWith,
  Subject,
  switchMap,
  takeUntil,
} from 'rxjs';
import {UuidComponent} from '../../components/uuid';
import {
  ArtifactLicense,
  ArtifactLicenseSelection,
  ArtifactLicensesService,
} from '../../services/artifact-licenses.service';
import {filteredByFormControl} from '../../../util/filter';
import {ApplicationsService} from '../../services/applications.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {getFormDisplayedError} from '../../../util/errors';
import {isExpired} from '../../../util/dates';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {EditLicenseComponent} from '../../licenses/edit-license.component';
import {dropdownAnimation} from '../../animations/dropdown';
import {drawerFlyInOut} from '../../animations/drawer';
import {modalFlyInOut} from '../../animations/modal';
import {EditArtifactLicenseComponent} from './edit-artifact-license.component';
import {UsersService} from '../../services/users.service';
import {ArtifactsService} from '../../services/artifacts.service';

@Component({
  selector: 'app-artifact-licenses',
  imports: [
    ReactiveFormsModule,
    AsyncPipe,
    FaIconComponent,
    UuidComponent,
    DatePipe,
    RequireRoleDirective,
    EditArtifactLicenseComponent,
  ],
  templateUrl: './artifact-licenses.component.html',
  animations: [dropdownAnimation, drawerFlyInOut, modalFlyInOut],
})
export class ArtifactLicensesComponent implements OnDestroy {
  private readonly destroyed$ = new Subject<void>();
  private readonly artifactLicensesService = inject(ArtifactLicensesService);

  filterForm = new FormGroup({
    search: new FormControl(''),
  });
  licenses$: Observable<ArtifactLicense[]> = filteredByFormControl(
    this.artifactLicensesService.list(),
    this.filterForm.controls.search,
    (it: ArtifactLicense, search: string) => !search || (it.name || '').toLowerCase().includes(search.toLowerCase())
  ).pipe(takeUntil(this.destroyed$));

  editForm = new FormGroup({
    license: new FormControl<ArtifactLicense | undefined>(undefined, {
      nonNullable: true,
      validators: Validators.required,
    }),
  });
  editFormLoading = false;

  private manageLicenseDrawerRef?: DialogRef;
  protected readonly faMagnifyingGlass = faMagnifyingGlass;

  private readonly overlay = inject(OverlayService);
  private readonly toast = inject(ToastService);
  private readonly usersService = inject(UsersService);
  private readonly users$ = this.usersService.getUsers();
  private readonly artifactsService = inject(ArtifactsService);
  private readonly artifacts$ = this.artifactsService.list();

  openDrawer(templateRef: TemplateRef<unknown>, license?: ArtifactLicense) {
    this.hideDrawer();
    if (license) {
      this.loadLicense(license);
    }
    this.manageLicenseDrawerRef = this.overlay.showDrawer(templateRef);
  }

  loadLicense(license: ArtifactLicense) {
    this.editForm.patchValue({license});
  }

  hideDrawer() {
    this.manageLicenseDrawerRef?.close();
    this.editForm.reset({license: undefined});
  }

  async saveLicense() {
    this.editForm.markAllAsTouched();
    const {license} = this.editForm.value;
    if (this.editForm.valid && license) {
      this.editFormLoading = true;
      const action = license.id
        ? this.artifactLicensesService.update(license)
        : this.artifactLicensesService.create(license);
      try {
        const license = await firstValueFrom(action);
        this.hideDrawer();
        this.toast.success(`${license.name} saved successfully`);
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.editFormLoading = false;
      }
    }
  }

  deleteLicense(license: ArtifactLicense) {
    this.overlay
      .confirm(`Really delete ${license.name}?`)
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.artifactLicensesService.delete(license)),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return EMPTY;
        })
      )
      .subscribe();
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  getArtifactColumn(selection?: ArtifactLicenseSelection[]): Observable<string | undefined> {
    return selection?.[0]?.artifactId
      ? this.artifacts$.pipe(
          map((artifacts) => artifacts.find((a) => a.id === selection?.[0]?.artifactId)),
          map((a) => a?.name + (selection?.length > 1 ? ' (+' + (selection.length - 1) + ')' : ''))
        )
      : EMPTY;
  }

  getOwnerColumn(userAccountId?: string): Observable<string | undefined> {
    return userAccountId
      ? this.users$.pipe(
          map((users) => users.find((u) => u.id === userAccountId)),
          map((u) => u?.name ?? u?.email)
        )
      : EMPTY;
  }

  protected readonly faPlus = faPlus;
  protected readonly isExpired = isExpired;
  protected readonly faPen = faPen;
  protected readonly faTrash = faTrash;
  protected readonly faXmark = faXmark;
}
