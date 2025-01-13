import {Component, inject, Input} from '@angular/core';
import {firstValueFrom} from 'rxjs';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DeploymentTarget} from '../../types/deployment-target';

@Component({
  selector: 'app-connect-instructions',
  templateUrl: './connect-instructions.component.html',
})
export class ConnectInstructionsComponent {
  @Input({required: true}) deploymentTarget!: DeploymentTarget;

  private readonly deploymentTargets = inject(DeploymentTargetsService);

  modalConnectCommand?: string;
  modalTargetId?: string;
  modalTargetSecret?: string;
  commandCopied = false;

  ngOnInit() {
    this.deploymentTargets.requestAccess(this.deploymentTarget.id!).subscribe((response) => {
      this.modalConnectCommand =
        this.deploymentTarget.type === 'docker'
          ? `curl "${response.connectUrl}" | docker compose -f - up -d`
          : `kubectl apply -n ${this.deploymentTarget.namespace} -f "${response.connectUrl}"`;
      this.modalTargetId = response.targetId;
      this.modalTargetSecret = response.targetSecret;
    });
  }

  async copyConnectCommand() {
    if (this.modalConnectCommand) {
      await navigator.clipboard.writeText(this.modalConnectCommand);
    }
    this.commandCopied = true;
  }
}
