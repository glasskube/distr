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
  modalTargetId?: string;
  modalTargetSecret?: string;
  commandCopied = false;

  ngOnInit() {
    this.deploymentTargets.requestAccess(this.deploymentTargetId).subscribe((response) => {
      this.modalConnectCommand = `curl "${response.connectUrl}" | docker compose -f - up -d`;
      this.modalTargetId = response.targetId;
      this.modalTargetSecret = response.targetSecret;
    });
  }

  async copyConnectCommand() {
    await navigator.clipboard.writeText(this.modalConnectCommand || '');
    this.commandCopied = true;
  }
}
