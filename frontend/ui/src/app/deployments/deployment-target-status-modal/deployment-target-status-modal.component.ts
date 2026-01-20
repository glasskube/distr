import {Component, input, output} from '@angular/core';
import {DeploymentTarget} from '@distr-sh/distr-sdk';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {DeploymentTargetLogsTableComponent} from './deployment-target-logs-table.component';

@Component({
  selector: 'app-deployment-target-status-modal',
  templateUrl: './deployment-target-status-modal.component.html',
  imports: [DeploymentTargetLogsTableComponent, FaIconComponent],
})
export class DeploymentTargetStatusModalComponent {
  public readonly deploymentTarget = input.required<DeploymentTarget>();
  public readonly closed = output<void>();

  protected readonly faXmark = faXmark;

  protected hideModal() {
    this.closed.emit();
  }
}
