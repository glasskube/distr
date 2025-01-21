import {OverlayModule} from '@angular/cdk/overlay';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Component, ElementRef, inject, Input, OnDestroy, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBoxArchive, faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {catchError, EMPTY, filter, firstValueFrom, Observable, of, Subject, switchMap, takeUntil} from 'rxjs';
import {drawerFlyInOut} from '../animations/drawer';
import {dropdownAnimation} from '../animations/dropdown';
import {modalFlyInOut} from '../animations/modal';
import {RequireRoleDirective} from '../directives/required-role.directive';
import {ApplicationsService} from '../services/applications.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';
import {Application, ApplicationVersion} from '../types/application';
import {filteredByFormControl} from '../../util/filter';
import {disableControlsWithoutEvent, enableControlsWithoutEvent} from '../../util/forms';
import {DeploymentType, HelmChartType} from '../types/deployment';
import {getFormDisplayedError} from '../../util/errors';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {YamlEditorComponent} from '../components/yaml-editor.component';

@Component({
  selector: 'app-applications',
  imports: [
    AsyncPipe,
    DatePipe,
    ReactiveFormsModule,
    FaIconComponent,
    NgOptimizedImage,
    OverlayModule,
    RequireRoleDirective,
    AutotrimDirective,
    YamlEditorComponent,
  ],
  templateUrl: './applications.component.html',
  animations: [dropdownAnimation, drawerFlyInOut, modalFlyInOut],
})
export class ApplicationsComponent implements OnInit, OnDestroy {
  @Input('fullVersion') fullVersion: boolean = false;
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
  selectedApplication?: Application;
  editForm = new FormGroup({
    id: new FormControl(''),
    name: new FormControl('', Validators.required),
    type: new FormControl<DeploymentType>('docker', Validators.required),
  });
  editFormLoading = false;
  newVersionForm = new FormGroup({
    versionName: new FormControl('', Validators.required),
    kubernetes: new FormGroup({
      chartType: new FormControl<HelmChartType>('repository', {
        nonNullable: true,
        validators: Validators.required,
      }),
      chartName: new FormControl<string>('', Validators.required),
      chartUrl: new FormControl<string>('', Validators.required),
      chartVersion: new FormControl<string>('', Validators.required),
      baseValues: new FormControl<string>(''),
      template: new FormControl<string>(''),
    }),
    docker: new FormGroup({
      compose: new FormControl<string>('', Validators.required),
    }),
  });
  newVersionFormLoading = false;

  private manageApplicationDrawerRef?: DialogRef;
  private applicationVersionModalRef?: DialogRef;

  private readonly overlay = inject(OverlayService);

  private readonly toast = inject(ToastService);

  ngOnInit() {
    this.newVersionForm.controls.kubernetes.controls.chartType.valueChanges
      .pipe(takeUntil(this.destroyed$))
      .subscribe((type) => {
        if (type === 'repository') {
          this.newVersionForm.controls.kubernetes.controls.chartName.enable();
        } else {
          this.newVersionForm.controls.kubernetes.controls.chartName.disable();
        }
      });
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  openDrawer(templateRef: TemplateRef<unknown>, application?: Application) {
    this.hideDrawer();
    if (application) {
      this.loadApplication(application);
    } else {
      this.reset();
    }
    this.manageApplicationDrawerRef = this.overlay.showDrawer(templateRef);
  }

  hideDrawer() {
    this.manageApplicationDrawerRef?.close();
    this.reset();
  }

  openVersionModal(templateRef: TemplateRef<unknown>, application: Application) {
    this.hideVersionModal();
    this.loadApplication(application);
    this.applicationVersionModalRef = this.overlay.showModal(templateRef);
  }

  hideVersionModal() {
    this.applicationVersionModalRef?.close();
    this.resetVersionForm();
  }

  loadApplication(application: Application) {
    this.selectedApplication = application;
    this.editForm.patchValue({
      id: application.id,
      type: application.type,
      name: application.name,
    });
    this.editForm.controls.type.disable();
    this.resetVersionForm();
    if (this.selectedApplication?.type === 'kubernetes') {
      enableControlsWithoutEvent(this.newVersionForm.controls.kubernetes);
      disableControlsWithoutEvent(this.newVersionForm.controls.docker);
    } else {
      enableControlsWithoutEvent(this.newVersionForm.controls.docker);
      disableControlsWithoutEvent(this.newVersionForm.controls.kubernetes);
    }
  }

  reset() {
    this.selectedApplication = undefined;
    this.resetEditForm();
    this.resetVersionForm();
  }

  deleteApplication(application: Application) {
    this.overlay
      .confirm(`Really delete ${application.name} and all related deployments?`)
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.applications.delete(application)),
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

  private resetEditForm() {
    this.editForm.reset();
    this.editForm.patchValue({type: 'docker'});
    this.editForm.controls.type.enable();
  }

  private resetVersionForm() {
    this.newVersionForm.reset();
  }

  async saveApplication() {
    this.editForm.markAllAsTouched();
    if (this.editForm.valid) {
      this.editFormLoading = true;
      const val = this.editForm.getRawValue();
      let result: Observable<Application>;
      if (!val.id) {
        result = this.applications.create({
          name: val.name!,
          type: val.type!,
        });
      } else {
        result = this.applications.update({
          id: val.id!,
          name: val.name!,
          type: val.type!,
        });
      }
      try {
        const application = await firstValueFrom(result);
        this.hideDrawer();
        this.toast.success(`${application.name} saved successfully`);
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

  async createVersion() {
    this.newVersionForm.markAllAsTouched();
    const isDocker = this.selectedApplication?.type === 'docker';
    if (this.newVersionForm.valid && this.selectedApplication) {
      this.newVersionFormLoading = true;
      let res;
      if (isDocker) {
        res = this.applications.createApplicationVersionForDocker(
          this.selectedApplication,
          {
            name: this.newVersionForm.controls.versionName.value!,
          },
          this.newVersionForm.controls.docker.controls.compose.value!
        );
      } else {
        const versionFormVal = this.newVersionForm.controls.kubernetes.value;
        const version = {
          name: this.newVersionForm.controls.versionName.value!,
          chartType: versionFormVal.chartType!,
          chartName: versionFormVal.chartName ?? undefined,
          chartUrl: versionFormVal.chartUrl!,
          chartVersion: versionFormVal.chartVersion!,
        };
        res = this.applications.createApplicationVersionForKubernetes(
          this.selectedApplication,
          version,
          versionFormVal.baseValues,
          versionFormVal.template
        );
      }

      try {
        const av = await firstValueFrom(res);
        this.toast.success(`${av.name} created successfully`);
        this.hideVersionModal();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.newVersionFormLoading = false;
      }
    }
  }

  async fillVersionFormWith(selectedApplication: Application | undefined, version: ApplicationVersion) {
    if (selectedApplication?.type === 'kubernetes') {
      try {
        const template = await firstValueFrom(this.applications.getTemplateFile(selectedApplication.id!, version.id!));
        const values = await firstValueFrom(this.applications.getValuesFile(selectedApplication.id!, version.id!));
        this.newVersionForm.patchValue({
          kubernetes: {
            chartType: version.chartType,
            chartName: version.chartName,
            chartUrl: version.chartUrl,
            chartVersion: version.chartVersion,
            baseValues: values,
            template: template,
          },
        });
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      }
    } else if (selectedApplication?.type === 'docker') {
      try {
        const compose = await firstValueFrom(this.applications.getComposeFile(selectedApplication.id!, version.id!));
        this.newVersionForm.patchValue({
          docker: {
            compose,
          },
        });
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      }
    }
  }
}
