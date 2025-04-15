import {AfterViewInit, Component, inject, OnDestroy, OnInit, signal, ViewChild} from '@angular/core';
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
import {Router} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {faCircleCheck} from '@fortawesome/free-regular-svg-icons';
import {firstValueFrom, lastValueFrom, Subject, takeUntil, tap} from 'rxjs';
import {AccessTokenWithKey} from '../../../../../../sdk/js/src';
import {getFormDisplayedError} from '../../../util/errors';
import {HttpErrorResponse} from '@angular/common/http';
import {ToastService} from '../../services/toast.service';
import {TutorialsService} from '../../services/tutorials.service';
import {TutorialProgress} from '../../types/tutorials';
import {Organization} from '../../types/organization';
import {OrganizationService} from '../../services/organization.service';
import {slugMaxLength, slugPattern} from '../../../util/slug';
import {fromPromise} from 'rxjs/internal/observable/innerFrom';
import {getRemoteEnvironment} from '../../../env/remote';
import {AccessTokensService} from '../../services/access-tokens.service';
import {ClipComponent} from '../../components/clip.component';

const tutorialId = 'registry';
const welcomeStep = 'welcome';
const welcomeTaskStart = 'start';
const prepareStep = 'prepare';
const prepareStepTaskSetSlug = 'set-slug';
const loginStep = 'login';
const loginStepTaskCreateToken = 'create-token';
const loginStepTaskLogin = 'login';
const usageStep = 'usage';
const usageStepTaskPull = 'pull';
const usageStepTaskTag = 'tag';
const usageStepTaskPush = 'push';
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
    ClipComponent,
  ],
  templateUrl: './registry-tutorial.component.html',
})
export class RegistryTutorialComponent implements OnInit, AfterViewInit, OnDestroy {
  loading = signal(true);
  private readonly destroyed$ = new Subject<void>();
  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;
  protected readonly faPalette = faPalette;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faB = faB;
  protected readonly faLightbulb = faLightbulb;
  @ViewChild('stepper', {static: false}) private stepper?: CdkStepper;
  private readonly router = inject(Router);
  protected readonly toast = inject(ToastService);
  protected readonly organizationService = inject(OrganizationService);
  protected readonly tutorialsService = inject(TutorialsService);
  protected readonly tokenService = inject(AccessTokensService);
  protected createdToken?: AccessTokenWithKey;
  protected progress?: TutorialProgress;
  private organization?: Organization;
  protected readonly welcomeFormGroup = new FormGroup({});
  protected readonly prepareFormGroup = new FormGroup({
    slugDone: new FormControl<boolean>(false),
    slug: new FormControl<string>('', {
      nonNullable: true,
      validators: [Validators.required, Validators.pattern(slugPattern), Validators.maxLength(slugMaxLength)],
    }),
  });
  protected readonly loginFormGroup = new FormGroup({
    tokenDone: new FormControl<boolean>(false, Validators.requiredTrue),
    loginDone: new FormControl<boolean>(false, Validators.requiredTrue),
  });
  protected readonly usageFormGroup = new FormGroup({
    pullDone: new FormControl<boolean>(false, Validators.requiredTrue),
    tagDone: new FormControl<boolean>(false, Validators.requiredTrue),
    pushDone: new FormControl<boolean>(false, Validators.requiredTrue),
    exploreDone: new FormControl<boolean>(false, Validators.requiredTrue),
  });
  protected readonly registrySlug$ = this.organizationService.get().pipe(tap((o) => (this.slug = o.slug)));
  protected slug?: string;
  protected readonly registryHost$ = fromPromise(getRemoteEnvironment()).pipe(tap((e) => (this.host = e.registryHost)));
  protected host?: string;

  ngOnInit() {
    this.usageFormGroup.controls.pullDone.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        tap((done) => this.saveDoneIfNotYetDone(done ?? false, usageStep, usageStepTaskPull))
      )
      .subscribe();
    this.usageFormGroup.controls.tagDone.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        tap((done) => this.saveDoneIfNotYetDone(done ?? false, usageStep, usageStepTaskTag))
      )
      .subscribe();
    this.usageFormGroup.controls.pushDone.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        tap((done) => this.saveDoneIfNotYetDone(done ?? false, usageStep, usageStepTaskPush))
      )
      .subscribe();
  }

  async ngAfterViewInit() {
    try {
      this.progress = await lastValueFrom(this.tutorialsService.get(tutorialId));
      if (this.progress.createdAt) {
        if (!this.progress.completedAt) {
          await this.continueFromWelcome();
          await this.continueFromPrepare();
          await this.continueFromLogin();
        } else {
          this.stepper?.steps.forEach((s) => (s.completed = true));
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

  private async saveDoneIfNotYetDone(done: boolean, stepId: string, taskId: string) {
    const doneBefore = (this.progress?.events ?? []).find((e) => e.stepId === stepId && e.taskId === taskId);
    if (done && !doneBefore) {
      this.progress = await firstValueFrom(
        this.tutorialsService.save(tutorialId, {
          stepId: stepId,
          taskId: taskId,
        })
      );
    }
  }

  protected async continueFromWelcome() {
    try {
      this.loading.set(true);
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
        this.stepper?.next();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.loading.set(false);
      }
    } else {
      this.stepper?.next();
    }
  }

  protected async continueFromPrepare() {
    this.prepareFormGroup.markAllAsTouched();
    if (this.prepareFormGroup.valid) {
      if (this.prepareFormGroup.dirty) {
        this.loading.set(true);
        const formVal = this.prepareFormGroup.getRawValue();
        try {
          this.organization = await lastValueFrom(
            this.organizationService.update({
              ...this.organization!,
              slug: formVal.slug!,
            })
          );
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
      this.prepareLoginStep();
      this.stepper?.next();
    }
  }

  private prepareLoginStep() {
    const token = (this.progress?.events ?? []).find(
      (e) => e.stepId === loginStep && e.taskId === loginStepTaskCreateToken
    );
    const login = (this.progress?.events ?? []).find((e) => e.stepId === loginStep && e.taskId === loginStepTaskLogin);

    this.loginFormGroup.patchValue({
      tokenDone: !!token,
      loginDone: !!login,
    });
  }

  protected async createToken() {
    this.loading.set(true);
    try {
      this.createdToken = await firstValueFrom(
        this.tokenService.create({
          label: 'tutorial-token',
          // TODO expiry?
        })
      );
      this.loginFormGroup.controls.tokenDone.patchValue(true);
      this.toast.success('Personal Access Token created successfully');
      this.progress = await lastValueFrom(
        this.tutorialsService.save(tutorialId, {
          stepId: loginStep,
          taskId: loginStepTaskCreateToken,
        })
      );
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    } finally {
      this.loading.set(false);
    }
  }

  protected async continueFromLogin() {
    this.loginFormGroup.markAllAsTouched();
    if (this.loginFormGroup.valid) {
      if (this.loginFormGroup.dirty) {
        this.loading.set(true);
        try {
          this.loginFormGroup.markAsPristine();
          this.progress = await lastValueFrom(
            this.tutorialsService.save(tutorialId, {
              stepId: loginStep,
              taskId: loginStepTaskLogin,
            })
          );
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

      this.prepareUsageStep();
      this.stepper?.next();
    }
  }

  private prepareUsageStep() {
    const pull = (this.progress?.events ?? []).find((e) => e.stepId === usageStep && e.taskId === usageStepTaskPull);
    const tag = (this.progress?.events ?? []).find((e) => e.stepId === usageStep && e.taskId === usageStepTaskTag);
    const push = (this.progress?.events ?? []).find((e) => e.stepId === usageStep && e.taskId === usageStepTaskPush);
    const explore = (this.progress?.events ?? []).find(
      (e) => e.stepId === usageStep && e.taskId === usageStepTaskExplore
    );

    this.usageFormGroup.patchValue({
      pullDone: !!pull,
      tagDone: !!tag,
      pushDone: !!push,
      exploreDone: !!explore,
    });
    this.usageFormGroup.markAsPristine();
  }

  protected async completeAndExit() {
    this.usageFormGroup.markAllAsTouched();
    if (this.usageFormGroup.valid && this.usageFormGroup.dirty) {
      this.loading.set(true);
      this.progress = await lastValueFrom(
        this.tutorialsService.save(tutorialId, {
          stepId: usageStep,
          taskId: usageStepTaskExplore,
          markCompleted: true,
        })
      );
      this.stepper!.selected!.completed = true;
      this.loading.set(false);
      this.toast.success('Congrats on finishing the tutorial! Good Job!');
      this.navigateToOverviewPage();
    } else if (this.progress?.completedAt) {
      this.navigateToOverviewPage();
    }
  }

  protected navigateToOverviewPage() {
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
