import {GlobalPositionStrategy, OverlayModule} from '@angular/cdk/overlay';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Component, inject, Input, OnDestroy, OnInit, TemplateRef} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBoxArchive, faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {catchError, EMPTY, filter, lastValueFrom, Observable, Subject, switchMap, takeUntil} from 'rxjs';
import {drawerFlyInOut} from '../animations/drawer';
import {dropdownAnimation} from '../animations/dropdown';
import {modalFlyInOut} from '../animations/modal';
import {ApplicationsService} from '../services/applications.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';
import {filteredByFormControl} from '../../util/filter';
import {getFormDisplayedError} from '../../util/errors';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {Application, DeploymentType} from '@glasskube/distr-sdk';
import {UuidComponent} from '../components/uuid';
import {RequireRoleDirective} from '../directives/required-role.directive';
import {Router, RouterLink} from '@angular/router';

@Component({
  selector: 'app-applications',
  imports: [
    AsyncPipe,
    DatePipe,
    ReactiveFormsModule,
    FaIconComponent,
    NgOptimizedImage,
    OverlayModule,
    AutotrimDirective,
    UuidComponent,
    RequireRoleDirective,
    RouterLink,
  ],
  templateUrl: './applications.component.html',
  animations: [dropdownAnimation, drawerFlyInOut, modalFlyInOut],
})
export class ApplicationsComponent implements OnInit, OnDestroy {
  @Input('fullVersion') fullVersion: boolean = false;
  private readonly router = inject(Router);
  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faPlus = faPlus;
  protected readonly faPen = faPen;
  protected readonly faXmark = faXmark;
  protected readonly faBoxArchive = faBoxArchive;
  protected readonly faTrash = faTrash;
  showDropdown = false;

  private readonly destroyed$ = new Subject<void>();
  private readonly applications = inject(ApplicationsService);
  filterForm = new FormGroup({
    search: new FormControl(''),
  });
  applications$: Observable<Application[]> = filteredByFormControl(
    this.applications.list(),
    this.filterForm.controls.search,
    (it: Application, search: string) => !search || (it.name || '').toLowerCase().includes(search.toLowerCase())
  ).pipe(takeUntil(this.destroyed$));
  editForm = new FormGroup({
    id: new FormControl(''),
    name: new FormControl('', Validators.required),
    type: new FormControl<DeploymentType>('docker', Validators.required),
  });
  editFormLoading = false;
  createApplicationForm = new FormGroup({
    name: new FormControl('', Validators.required),
    type: new FormControl<DeploymentType>('docker', Validators.required),
  });
  createApplicationFormLoading = false;

  private applicationCreateModalRef?: DialogRef;

  private readonly overlay = inject(OverlayService);
  private readonly toast = inject(ToastService);

  ngOnInit() {}

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  openCreateModal(templateRef: TemplateRef<unknown>) {
    this.hideCreateModal();
    this.applicationCreateModalRef = this.overlay.showModal(templateRef, {
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
    });
  }

  hideCreateModal() {
    this.applicationCreateModalRef?.close();
    this.createApplicationForm.reset();
  }

  async createApplication() {
    this.createApplicationForm.markAllAsTouched();
    if (this.createApplicationForm.valid) {
      this.createApplicationFormLoading = true;
      try {
        const formVal = this.createApplicationForm.value;
        const created = await lastValueFrom(
          this.applications.create({
            name: formVal.name!,
            type: formVal.type!,
          })
        );
        this.toast.success(`${formVal.name} created successfully`);
        this.hideCreateModal();
        await this.router.navigate(['/applications', created.id]);
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.createApplicationFormLoading = false;
      }
    }
  }
}
