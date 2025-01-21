import {Component, inject, Input} from '@angular/core';
import {displayedInToast, getFormDisplayedError} from '../../../util/errors';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {ToastService} from '../../services/toast.service';
import {DeploymentTarget} from '../../types/deployment-target';
import {faClipboard} from '@fortawesome/free-regular-svg-icons';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faClipboardCheck} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-connect-instructions',
  templateUrl: './connect-instructions.component.html',
  imports: [FaIconComponent],
})
export class ConnectInstructionsComponent {
  @Input({required: true}) deploymentTarget!: DeploymentTarget;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly toast = inject(ToastService);

  modalConnectCommand?: string;
  modalTargetId?: string;
  modalTargetSecret?: string;
  commandCopied = false;

  ngOnInit() {
    this.deploymentTargets.requestAccess(this.deploymentTarget.id!).subscribe(
      (response) => {
        this.modalConnectCommand =
          this.deploymentTarget.type === 'docker'
            ? `curl "${response.connectUrl}" | docker compose -f - up -d`
            : `kubectl apply -n ${this.deploymentTarget.namespace} -f "${response.connectUrl}"`;
        this.modalTargetId = response.targetId;
        this.modalTargetSecret = response.targetSecret;
      },
      (e) => {
        if (!displayedInToast(e)) {
          this.toast.error(getFormDisplayedError(e) ?? e);
        }
      }
    );
  }

  async copyConnectCommand() {
    if (this.modalConnectCommand) {
      await navigator.clipboard.writeText(this.modalConnectCommand);
    }
    this.commandCopied = true;
  }

  protected readonly faClipboard = faClipboard;
  protected readonly faClipboardCheck = faClipboardCheck;
}
