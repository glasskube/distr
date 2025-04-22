import {AfterViewInit, Component, inject, OnDestroy, OnInit, signal, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {
  faArrowRight,
  faB,
  faBox,
  faBoxesStacked,
  faCheck,
  faClipboardCheck,
  faDownload,
  faLightbulb,
  faPalette,
  faRightToBracket,
} from '@fortawesome/free-solid-svg-icons';
import {CdkStep, CdkStepper, CdkStepperPrevious} from '@angular/cdk/stepper';
import {TutorialStepperComponent} from '../stepper/tutorial-stepper.component';
import {Router} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faCircleCheck, faClipboard} from '@fortawesome/free-regular-svg-icons';
import {firstValueFrom, lastValueFrom, Observable, OperatorFunction, Subject, switchMap, takeUntil, tap} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {HttpErrorResponse} from '@angular/common/http';
import {ToastService} from '../../services/toast.service';
import {TutorialsService} from '../../services/tutorials.service';
import {TutorialProgress} from '../../types/tutorials';
import {getExistingTask} from '../utils';
import {ApplicationsService} from '../../services/applications.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {ClipComponent} from '../../components/clip.component';

const tutorialId = 'agents';
const welcomeStep = 'welcome';
const welcomeTaskStart = 'start';
const deployStep = 'deploy';
const deployStepTaskDeploy = 'deploy';
const deployStepTaskVerify = 'verify';
const releaseStep = 'release';
const releaseStepTaskFork = 'fork';
const releaseStepTaskPAT = 'pat';
const releaseStepTaskCopyID = 'copy-id';
const releaseStepTaskRelease = 'release';
const releaseStepTaskVerify = 'verify';

@Component({
  selector: 'app-agents-tutorial',
  imports: [ReactiveFormsModule, CdkStep, TutorialStepperComponent, FaIconComponent, CdkStepperPrevious, ClipComponent],
  templateUrl: './agents-tutorial.component.html',
})
export class AgentsTutorialComponent implements OnInit, AfterViewInit, OnDestroy {
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
  protected readonly applicationsService = inject(ApplicationsService);
  protected readonly tutorialsService = inject(TutorialsService);
  private readonly deploymentTargetService = inject(DeploymentTargetsService);
  protected progress?: TutorialProgress;
  protected readonly welcomeFormGroup = new FormGroup({});
  protected readonly deployFormGroup = new FormGroup({
    deployDone: new FormControl<boolean>(false, Validators.requiredTrue),
    verifyDone: new FormControl<boolean>(false, Validators.requiredTrue),
  });
  protected readonly releaseFormGroup = new FormGroup({
    forkDone: new FormControl<boolean>(false, Validators.requiredTrue),
    patDone: new FormControl<boolean>(false, Validators.requiredTrue),
    copyIdDone: new FormControl<boolean>(false, Validators.requiredTrue),
    // TODO note/step to update the deployed docker compose yaml
    // TODO enable github actions
    releaseDone: new FormControl<boolean>(false, Validators.requiredTrue),
    verifyDone: new FormControl<boolean>(false, Validators.requiredTrue),
  });
  connectCommand?: string;
  targetId?: string;
  targetSecret?: string;
  commandCopied = false;

  ngOnInit() {
    this.registerTaskToggle(this.deployFormGroup.controls.deployDone, deployStep, deployStepTaskDeploy);
    this.registerTaskToggle(this.deployFormGroup.controls.verifyDone, deployStep, deployStepTaskVerify);

    this.registerTaskToggle(this.releaseFormGroup.controls.forkDone, releaseStep, releaseStepTaskFork);
    this.registerTaskToggle(this.releaseFormGroup.controls.patDone, releaseStep, releaseStepTaskPAT);
    this.registerTaskToggle(this.releaseFormGroup.controls.copyIdDone, releaseStep, releaseStepTaskCopyID);
    this.registerTaskToggle(this.releaseFormGroup.controls.releaseDone, releaseStep, releaseStepTaskRelease);
    this.registerTaskToggle(this.releaseFormGroup.controls.verifyDone, releaseStep, releaseStepTaskVerify);
  }

  private registerTaskToggle(ctrl: FormControl<boolean | null>, stepId: string, taskId: string) {
    ctrl.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        switchMap((done) =>
          this.tutorialsService.saveDoneIfNotYetDone(this.progress, done ?? false, tutorialId, stepId, taskId)
        ),
        tap((updated) => (this.progress = updated))
      )
      .subscribe();
  }

  async ngAfterViewInit() {
    try {
      this.progress = await lastValueFrom(this.tutorialsService.get(tutorialId));
      if (this.progress.createdAt) {
        if (!this.progress.completedAt) {
          await this.continueFromWelcome();
          await this.continueFromDeploy();
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
    } finally {
      this.loading.set(false);
    }
  }

  protected async continueFromWelcome() {
    if (!this.progress) {
      this.loading.set(true);
      try {
        this.progress = await lastValueFrom(
          this.tutorialsService.save(tutorialId, {
            stepId: welcomeStep,
            taskId: welcomeTaskStart,
          })
        );
        this.applicationsService.refresh();
        this.stepper?.next();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
          return;
        }
      } finally {
        this.loading.set(false);
      }
    } else {
      this.stepper?.next();
    }

    this.prepareDeployStep();
  }

  private prepareDeployStep() {
    const deployed = getExistingTask(this.progress, deployStep, deployStepTaskDeploy);
    const verified = getExistingTask(this.progress, deployStep, deployStepTaskVerify);
    this.deployFormGroup.patchValue({
      deployDone: !!deployed,
      verifyDone: !!verified,
    });
  }

  protected async requestAccess() {
    const startTask = getExistingTask(this.progress, welcomeStep, welcomeTaskStart);
    if (startTask && 'deploymentTargetId' in startTask.value) {
      try {
        this.loading.set(true);
        const resp = await firstValueFrom(
          this.deploymentTargetService.requestAccess(startTask.value['deploymentTargetId'])
        );
        this.connectCommand = `curl "${resp.connectUrl}" | docker compose -f - up -d`;
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.loading.set(false);
      }
    }
  }

  async copyConnectCommand() {
    if (this.connectCommand) {
      await navigator.clipboard.writeText(this.connectCommand);
    }
    this.commandCopied = true;
    setTimeout(() => (this.commandCopied = false), 2000);
  }

  protected async continueFromDeploy() {
    this.deployFormGroup.markAllAsTouched();
    if (this.deployFormGroup.valid) {
      this.prepareReleaseStep();
      this.stepper?.next();
    }
  }

  private prepareReleaseStep() {
    const fork = getExistingTask(this.progress, releaseStep, releaseStepTaskFork);
    const pat = getExistingTask(this.progress, releaseStep, releaseStepTaskPAT);
    const copy = getExistingTask(this.progress, releaseStep, releaseStepTaskCopyID);
    const release = getExistingTask(this.progress, releaseStep, releaseStepTaskRelease);
    const verify = getExistingTask(this.progress, releaseStep, releaseStepTaskVerify);
    this.releaseFormGroup.patchValue({
      forkDone: !!fork,
      patDone: !!pat,
      copyIdDone: !!copy,
      releaseDone: !!release,
      verifyDone: !!verify,
    });
  }

  protected async completeAndExit() {
    this.releaseFormGroup.markAllAsTouched();
    if (this.releaseFormGroup.valid && this.releaseFormGroup.dirty) {
      this.loading.set(true);
      this.progress = await lastValueFrom(
        this.tutorialsService.save(tutorialId, {
          stepId: releaseStep,
          taskId: releaseStepTaskVerify,
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

  protected readonly faClipboard = faClipboard;
  protected readonly faClipboardCheck = faClipboardCheck;
}
