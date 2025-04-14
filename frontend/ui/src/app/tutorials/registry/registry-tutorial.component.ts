import {Component, inject, OnDestroy, OnInit, signal, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {
  faArrowRight,
  faB,
  faBox,
  faBoxesStacked,
  faCheck,
  faDownload,
  faLightbulb,
  faPalette,
  faRightToBracket,
} from '@fortawesome/free-solid-svg-icons';
import {CdkStep, CdkStepper, CdkStepperPrevious} from '@angular/cdk/stepper';
import {TutorialStepperComponent} from '../stepper/tutorial-stepper.component';
import {OrganizationBrandingService} from '../../services/organization-branding.service';
import {Router, RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {faCircleCheck} from '@fortawesome/free-regular-svg-icons';
import {firstValueFrom, lastValueFrom, map, Observable, Subject, takeUntil, tap} from 'rxjs';
import {OrganizationBranding} from '../../../../../../sdk/js/src';
import {base64ToBlob} from '../../../util/blob';
import {getFormDisplayedError} from '../../../util/errors';
import {HttpErrorResponse} from '@angular/common/http';
import {ToastService} from '../../services/toast.service';
import {UsersService} from '../../services/users.service';
import {TutorialsService} from '../../services/tutorials.service';
import {TutorialProgress} from '../../types/tutorials';
import {AuthService} from '../../services/auth.service';
import {Organization} from '../../types/organization';
import {OrganizationService} from '../../services/organization.service';

const tutorialId = 'registry';
const welcomeStep = 'welcome';
const welcomeTaskStart = 'start';
const prepareStep = 'prepare';
const prepareStepTaskSetSlug = 'set-slug';
const usageStep = 'usage';
const usageStepTaskLogin = 'login';
const usageStepTaskPull = 'pull';
const usageStepTaskExplore = 'explore';

@Component({
  selector: 'app-registry-tutorial',
  imports: [
    ReactiveFormsModule,
    CdkStep,
    TutorialStepperComponent,
    FaIconComponent,
    CdkStepperPrevious,
    AutotrimDirective,
  ],
  templateUrl: './registry-tutorial.component.html',
})
export class RegistryTutorialComponent implements OnInit, OnDestroy {
  loading = signal(true);
  private readonly destroyed$ = new Subject<void>();
  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;
  protected readonly faPalette = faPalette;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faB = faB;
  protected readonly faLightbulb = faLightbulb;
  @ViewChild('stepper') private stepper!: CdkStepper;
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);
  protected readonly toast = inject(ToastService);
  protected readonly organizationService = inject(OrganizationService);
  protected readonly usersService = inject(UsersService);
  protected readonly tutorialsService = inject(TutorialsService);
  protected progress?: TutorialProgress;
  private organization?: Organization;
  protected readonly welcomeFormGroup = new FormGroup({});
  protected readonly prepareFormGroup = new FormGroup({
    slugDone: new FormControl<boolean>(false),
    slug: new FormControl<string>('', {nonNullable: true, validators: Validators.required}),
  });
  protected readonly usageFormGroup = new FormGroup({
    // TODO instant save for these checkboxes??
    loginDone: new FormControl<boolean>(false, Validators.requiredTrue),
    pullDone: new FormControl<boolean>(false, Validators.requiredTrue),
    exploreDone: new FormControl<boolean>(false, Validators.requiredTrue),
  });

  async ngOnInit() {
    try {
      this.progress = await lastValueFrom(this.tutorialsService.get(tutorialId));
      if (this.progress.createdAt) {
        if (!this.progress.completedAt) {
          await this.continueFromWelcome();
          await this.continueFromPrepare();
          // TODO ?
        } else {
          this.stepper.steps.forEach((s) => (s.completed = true));
        }
      }
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg && e instanceof HttpErrorResponse && e.status !== 404) {
        // it's a valid use case for a tutorial progress not to exist yet
        this.toast.error(msg);
      }
    }
  }

  protected async continueFromWelcome() {
    // TODO prepare hello-distr images around here somehow
    try {
      this.organization = await firstValueFrom(this.organizationService.get());
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
      return;
    } finally {
      this.loading.set(false);
    }

    this.prepareFormGroup.patchValue({
      slug: this.organization?.slug,
      slugDone: !!this.organization?.slug,
    });

    if (!this.progress) {
      this.loading.set(true);
      try {
        this.progress = await lastValueFrom(
          this.tutorialsService.save(tutorialId, {
            stepId: welcomeStep,
            taskId: welcomeTaskStart,
          })
        );
        this.stepper.next();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.loading.set(false);
      }
    } else {
      this.stepper.next();
    }
  }

  protected async continueFromPrepare() {
    this.prepareFormGroup.markAllAsTouched();
    if (this.prepareFormGroup.valid) {
      if (this.prepareFormGroup.dirty) {
        this.loading.set(true);
        const formVal = this.prepareFormGroup.getRawValue();
        try {
          this.organization = await lastValueFrom(this.organizationService.update({
            ...this.organization!,
            slug: formVal.slug!,
          }));
          this.prepareFormGroup.markAsPristine();
          this.progress = await lastValueFrom(
            this.tutorialsService.save(tutorialId, {
              stepId: prepareStep,
              taskId: prepareStepTaskSetSlug,
            })
          );
          this.toast.success('Organization has been updated');
        } catch (e) {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return;
        } finally {
          this.loading.set(false);
        }
      }

      this.prepareFormGroup.controls.slugDone.patchValue(true);
      this.prepareUsageStep();
      this.stepper.next();
    }
  }

  private prepareUsageStep() {
    const login = (this.progress?.events ?? []).find(
      (e) => e.stepId === usageStep && e.taskId === usageStepTaskLogin
    );
    if (login) {
      this.usageFormGroup.controls.loginDone.patchValue(true);
    }

    const pull = (this.progress?.events ?? []).find(
      (e) => e.stepId === usageStep && e.taskId === usageStepTaskPull
    );
    if (pull) {
      this.usageFormGroup.controls.pullDone.patchValue(true);
    }

    const explore = (this.progress?.events ?? []).find(
      (e) => e.stepId === usageStep && e.taskId === usageStepTaskExplore
    );
    if (explore) {
      this.usageFormGroup.controls.exploreDone.patchValue(true);
    }

    this.usageFormGroup.markAsPristine();
  }

  protected async completeAndExit() {
    this.usageFormGroup.markAllAsTouched();
    if (this.usageFormGroup.valid && this.usageFormGroup.dirty) {
      this.loading.set(true);
      // TODO maybe change to instant save
      this.progress = await lastValueFrom(
        this.tutorialsService.save(tutorialId, {
          stepId: usageStep,
          taskId: usageStepTaskExplore,
          markCompleted: true,
        })
      );
      this.stepper.selected!.completed = true;
      this.loading.set(false);
      this.toast.success('Congrats on finishing the tutorial! Good Job!');
      this.navigateToOverviewPage();
    } else if (this.progress?.completedAt) {
      this.navigateToOverviewPage();
    }
  }

  protected navigateToOverviewPage() {
    this.tutorialsService.refreshList();
    this.router.navigate(['tutorials']);
  }

  protected readonly faArrowRight = faArrowRight;
  protected readonly faRightToBracket = faRightToBracket;
  protected readonly faCheck = faCheck;
  protected readonly faCircleCheck = faCircleCheck;

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }
}
