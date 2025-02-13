import {Component, inject, OnDestroy, OnInit, TemplateRef} from '@angular/core';
import {AsyncPipe, DatePipe} from '@angular/common';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {catchError, EMPTY, filter, firstValueFrom, Observable, Subject, switchMap, takeUntil, tap} from 'rxjs';
import {filteredByFormControl} from '../../util/filter';
import {LicensesService} from '../services/licenses.service';
import {ApplicationLicense} from '../types/application-license';
import {UuidComponent} from '../components/uuid';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {getFormDisplayedError} from '../../util/errors';
import {ToastService} from '../services/toast.service';
import {RequireRoleDirective} from '../directives/required-role.directive';
import {dropdownAnimation} from '../animations/dropdown';
import {drawerFlyInOut} from '../animations/drawer';
import {modalFlyInOut} from '../animations/modal';
import {ApplicationsService} from '../services/applications.service';
import {EditLicenseComponent} from './edit-license.component';
import {AccessToken} from '../../../../../sdk/js/src';
import dayjs from 'dayjs';
import {isExpired} from '../../util/dates';

@Component({
  selector: 'app-licenses',
  templateUrl: './licenses.component.html',
  imports: [
    AsyncPipe,
    AutotrimDirective,
    ReactiveFormsModule,
    FaIconComponent,
    UuidComponent,
    DatePipe,
    RequireRoleDirective,
    EditLicenseComponent,
  ],
  animations: [dropdownAnimation, drawerFlyInOut, modalFlyInOut],
})
export class LicensesComponent implements OnDestroy {
  private readonly destroyed$ = new Subject<void>();
  private readonly licensesService = inject(LicensesService);
  private readonly applicationsService = inject(ApplicationsService);

  filterForm = new FormGroup({
    search: new FormControl(''),
  });
  licenses$: Observable<ApplicationLicense[]> = filteredByFormControl(
    this.licensesService.list(),
    this.filterForm.controls.search,
    (it: ApplicationLicense, search: string) => !search || (it.name || '').toLowerCase().includes(search.toLowerCase())
  ).pipe(takeUntil(this.destroyed$));
  applications$ = this.applicationsService.list();

  editForm = new FormGroup({
    license: new FormControl<ApplicationLicense | undefined>(undefined, {
      nonNullable: true,
      validators: Validators.required,
    }),
  });
  editFormLoading = false;

  private manageLicenseDrawerRef?: DialogRef;
  protected readonly faMagnifyingGlass = faMagnifyingGlass;

  private readonly overlay = inject(OverlayService);
  private readonly toast = inject(ToastService);

  openDrawer(templateRef: TemplateRef<unknown>, license?: ApplicationLicense) {
    this.hideDrawer();
    if (license) {
      this.loadLicense(license);
    }
    this.manageLicenseDrawerRef = this.overlay.showDrawer(templateRef);
  }

  loadLicense(license: ApplicationLicense) {
    this.editForm.patchValue({license});
  }

  hideDrawer() {
    this.manageLicenseDrawerRef?.close();
    this.editForm.reset({license: undefined});
  }

  async saveLicense() {
    this.editForm.markAllAsTouched();
    const {license} = this.editForm.value;
    console.log('save', license);
    if (this.editForm.valid && license) {
      this.editFormLoading = true;
      const action = license.id ? this.licensesService.update(license) : this.licensesService.create(license);
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

  deleteLicense(license: ApplicationLicense) {
    this.overlay
      .confirm(`Really delete ${license.name}?`)
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.licensesService.delete(license)),
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

  protected readonly faPlus = faPlus;
  protected readonly faXmark = faXmark;
  protected readonly faPen = faPen;
  protected readonly faTrash = faTrash;
  protected readonly isExpired = isExpired;
}
