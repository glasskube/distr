import {Component, inject, Input} from '@angular/core';
import {firstValueFrom} from 'rxjs';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';

@Component({
  selector: 'app-connect-instructions',
  templateUrl: './connect-instructions.component.html',
})
export class ConnectInstructionsComponent {
  @Input('deploymentTargetId') deploymentTargetId!: string;

  private readonly deploymentTargets = inject(DeploymentTargetsService);

  modalConnectCommand?: string;
  modalAccessKeyId?: string;
  modalAccessKeySecret?: string;
  commandCopied = false;

  ngOnInit() {
    this.deploymentTargets.requestAccess(this.deploymentTargetId).subscribe((response) => {
      this.modalConnectCommand = `curl "${response.connectUrl}" | docker compose -f - up -d`;
      this.modalAccessKeyId = response.accessKeyId;
      this.modalAccessKeySecret = response.accessKeySecret;
    });
  }

  async copyConnectCommand() {
    await navigator.clipboard.writeText(this.modalConnectCommand || '');
    this.commandCopied = true;
  }
}
