import {OverlayModule} from '@angular/cdk/overlay';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Component, ElementRef, inject, Input, OnDestroy, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBoxArchive, faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {filter, Observable, Subject, switchMap, takeUntil} from 'rxjs';
import {drawerFlyInOut} from '../animations/drawer';
import {dropdownAnimation} from '../animations/dropdown';
import {modalFlyInOut} from '../animations/modal';
import {RequireRoleDirective} from '../directives/required-role.directive';
import {ApplicationsService} from '../services/applications.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';
import {Application} from '../types/application';
import {filteredByFormControl} from '../../util/filter';
import {disableControls, enableControls} from '../../util/forms';
import {DeploymentType, HelmChartType} from '../types/deployment';

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
  newVersionForm = new FormGroup({
    versionName: new FormControl('', Validators.required),
    kubernetes: new FormGroup({
      chartType: new FormControl<HelmChartType>(
        {
          value: 'repository',
          disabled: true,
        },
        {
          nonNullable: true,
          validators: Validators.required,
        }
      ),
      chartName: new FormControl<string>(
        {
          value: '',
          disabled: true,
        },
        Validators.required
      ),
      chartUrl: new FormControl<string>(
        {
          value: '',
          disabled: true,
        },
        Validators.required
      ),
      chartVersion: new FormControl<string>(
        {
          value: '',
          disabled: true,
        },
        Validators.required
      ),
    }),
  });
  dockerComposeFile: File | null = null;
  @ViewChild('dockerComposeFileInput')
  dockerComposeFileInput?: ElementRef;

  private manageApplicationDrawerRef?: DialogRef;
  private applicationVersionModalRef?: DialogRef;

  baseValuesFile: File | null = null;
  @ViewChild('baseValuesFileInput')
  baseValuesFileInput?: ElementRef;

  templateFile: File | null = null;
  @ViewChild('templateFileInput')
  templateFileInput?: ElementRef;

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
      enableControls(this.newVersionForm.controls.kubernetes);
    } else {
      disableControls(this.newVersionForm.controls.kubernetes);
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
        switchMap(() => this.applications.delete(application))
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
    this.dockerComposeFile = null;
    if (this.dockerComposeFileInput) {
      this.dockerComposeFileInput.nativeElement.value = '';
    }
    this.baseValuesFile = null;
    if (this.baseValuesFileInput) {
      this.baseValuesFileInput.nativeElement.value = '';
    }
    this.templateFile = null;
    if (this.templateFileInput) {
      this.templateFileInput.nativeElement.value = '';
    }
  }

  saveApplication() {
    if (this.editForm.valid) {
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
      result.subscribe({
        next: (application) => {
          this.hideDrawer();
          this.toast.success(`${application.name} saved successfully`);
        },
        error: () => this.toast.error(`An error occurred`),
      });
    } else {
      this.editForm.markAllAsTouched();
    }
  }

  onDockerComposeSelected(event: any) {
    this.dockerComposeFile = event.target.files[0];
  }

  onBaseValuesFileSelected(event: Event) {
    this.baseValuesFile = (event.target as HTMLInputElement).files?.[0] ?? null;
  }

  onTemplateFileSelected(event: Event) {
    this.templateFile = (event.target as HTMLInputElement).files?.[0] ?? null;
  }

  createVersion() {
    const isDocker = this.selectedApplication?.type === 'docker';
    const fileValid = !isDocker || (isDocker && this.dockerComposeFile != null);
    if (this.newVersionForm.valid && fileValid && this.selectedApplication) {
      let res;
      if (isDocker) {
        res = this.applications.createApplicationVersionForDocker(
          this.selectedApplication,
          {
            name: this.newVersionForm.controls.versionName.value!,
          },
          this.dockerComposeFile!
        );
      } else {
        res = this.applications.createApplicationVersionForKubernetes(
          this.selectedApplication,
          {
            name: this.newVersionForm.controls.versionName.value!,
            chartType: this.newVersionForm.controls.kubernetes.controls.chartType.value,
            chartName:
              this.newVersionForm.controls.kubernetes.controls.chartType.value === 'repository'
                ? this.newVersionForm.controls.kubernetes.controls.chartName.value!
                : undefined,
            chartUrl: this.newVersionForm.controls.kubernetes.controls.chartUrl.value!,
            chartVersion: this.newVersionForm.controls.kubernetes.controls.chartVersion.value!,
          },
          this.baseValuesFile,
          this.templateFile
        );
      }
      res.subscribe((value) => {
        this.toast.success(`${value.name} created successfully`);
        this.hideVersionModal();
      });
    } else {
      this.newVersionForm.markAllAsTouched();
    }
  }
}
