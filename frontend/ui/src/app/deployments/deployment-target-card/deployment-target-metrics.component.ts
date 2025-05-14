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
  styles: `
    .gauge {
      position: relative;
      width: 40px;
      aspect-ratio: 1;
      border-radius: 50%;
      background: conic-gradient(green 0%, yellow 50%, red 100%);

      @for $i from 0 through 100 {
        &.percent-#{$i} {
          $deg: ($i * 3.6);
          mask: conic-gradient(white 0deg #{$deg}deg, #00000036 #{$deg}deg 360deg);
        }
      }
    }

    .gauge-center {
      position: absolute;
      top: 10%;
      left: 10%;
      width: 80%;
      height: 80%;
      border-radius: 50%;
    }
  `,
})
export class DeploymentTargetMetricsComponent {
  public readonly fullVersion = input(false);
  public readonly metrics = input.required<DeploymentTargetLatestMetrics>();

  protected getPercent(usage: number | undefined): string {
    return Math.ceil((usage || 0) * 100).toFixed();
  }
}
