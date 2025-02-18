import {OverlayModule} from '@angular/cdk/overlay';
import {Component, inject, OnDestroy, OnInit, signal} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute, RouterLink} from '@angular/router';
import {combineLatestWith, firstValueFrom, map, Observable, Subject, takeUntil, tap} from 'rxjs';
import {ApplicationsService} from '../services/applications.service';
import {AsyncPipe, DatePipe, JsonPipe, NgOptimizedImage} from '@angular/common';
import {Application, ApplicationVersion, HelmChartType} from '@glasskube/distr-sdk';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBoxesStacked, faChevronDown} from '@fortawesome/free-solid-svg-icons';
import {UuidComponent} from '../components/uuid';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {YamlEditorComponent} from '../components/yaml-editor.component';
import {getFormDisplayedError} from '../../util/errors';
import {ToastService} from '../services/toast.service';
import {disableControlsWithoutEvent, enableControlsWithoutEvent} from '../../util/forms';
import {dropdownAnimation} from '../animations/dropdown';

@Component({
  selector: 'app-application-detail',
  imports: [
    ReactiveFormsModule,
    OverlayModule,
    AsyncPipe,
    RouterLink,
    FaIconComponent,
    NgOptimizedImage,
    UuidComponent,
    AutotrimDirective,
    DatePipe,
    YamlEditorComponent,
  ],
  templateUrl: './application-detail.component.html',
  animations: [dropdownAnimation],
})
export class ApplicationDetailComponent implements OnInit, OnDestroy {
  private readonly destroyed$ = new Subject<void>();
  private readonly toast = inject(ToastService);
  private readonly applicationService = inject(ApplicationsService);
  private readonly route = inject(ActivatedRoute);
  readonly applications$: Observable<Application[]> = this.applicationService.list();
  readonly application$: Observable<Application | undefined> = this.route.paramMap.pipe(
    combineLatestWith(this.applications$),
    map(([params, applications]) => {
      const id = params.get('applicationId');
      return applications.find((a) => a.id === id);
    }),
    tap((app) => {
      this.newVersionForm.reset();
      if (app) {
        if (app.type === 'kubernetes') {
          enableControlsWithoutEvent(this.newVersionForm.controls.kubernetes);
          disableControlsWithoutEvent(this.newVersionForm.controls.docker);
        } else {
          enableControlsWithoutEvent(this.newVersionForm.controls.docker);
          disableControlsWithoutEvent(this.newVersionForm.controls.kubernetes);
        }
      }
    })
  );
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
      template: new FormControl<string>(''),
    }),
  });
  newVersionFormLoading = false;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faChevronDown = faChevronDown;
  readonly dropdownOpen = signal(false);

  ngOnInit() {
    this.route.url.pipe().subscribe(() => this.dropdownOpen.set(false));
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

  toggleDropdown() {
    this.dropdownOpen.update((v) => !v);
  }

  async createVersion() {
    this.newVersionForm.markAllAsTouched();
    const application = await firstValueFrom(this.application$);
    if (this.newVersionForm.valid && application) {
      this.newVersionFormLoading = true;
      let res;
      if (application.type === 'docker') {
        res = this.applicationService.createApplicationVersionForDocker(
          application,
          {
            name: this.newVersionForm.controls.versionName.value!,
          },
          this.newVersionForm.controls.docker.controls.compose.value!,
          this.newVersionForm.controls.docker.controls.template.value!
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
        res = this.applicationService.createApplicationVersionForKubernetes(
          application,
          version,
          versionFormVal.baseValues,
          versionFormVal.template
        );
      }

      try {
        const av = await firstValueFrom(res);
        this.toast.success(`${av.name} created successfully`);
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

  async fillVersionFormWith(version: ApplicationVersion) {
    const application = await firstValueFrom(this.application$);
    if (!application) {
      return;
    }
    if (application.type === 'kubernetes') {
      try {
        const template = await firstValueFrom(this.applicationService.getTemplateFile(application.id!, version.id!));
        const values = await firstValueFrom(this.applicationService.getValuesFile(application.id!, version.id!));
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
    } else if (application.type === 'docker') {
      try {
        const template = await firstValueFrom(this.applicationService.getTemplateFile(application.id!, version.id!));
        const compose = await firstValueFrom(this.applicationService.getComposeFile(application.id!, version.id!));
        this.newVersionForm.patchValue({
          docker: {
            compose,
            template: template,
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
