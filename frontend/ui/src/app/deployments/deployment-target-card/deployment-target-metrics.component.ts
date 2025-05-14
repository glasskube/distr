import {Component, input} from '@angular/core';
import {OverlayModule} from '@angular/cdk/overlay';
import {ReactiveFormsModule} from '@angular/forms';
import {modalFlyInOut} from '../../animations/modal';
import {drawerFlyInOut} from '../../animations/drawer';
import {dropdownAnimation} from '../../animations/dropdown';
import {DeploymentTargetLatestMetrics} from '../../services/deployment-target-metrics.service';
import {BytesPipe} from '../../../util/units';

@Component({
  selector: 'app-deployment-target-metrics',
  templateUrl: './deployment-target-metrics.component.html',
  imports: [OverlayModule, ReactiveFormsModule, BytesPipe],
  animations: [modalFlyInOut, drawerFlyInOut, dropdownAnimation],
  styleUrls: ['./deployment-target-metrics.component.scss'],
})
export class DeploymentTargetMetricsComponent {
  public readonly fullVersion = input(false);
  public readonly metrics = input.required<DeploymentTargetLatestMetrics>();

  protected getPercent(usage: number | undefined): string {
    return Math.ceil((usage || 0) * 100).toFixed();
  }
}
