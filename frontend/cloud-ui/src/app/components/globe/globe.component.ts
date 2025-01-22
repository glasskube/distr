import {
  ApplicationRef,
  Component,
  ElementRef,
  inject,
  Input,
  OnChanges,
  OnDestroy,
  OnInit,
  SimpleChanges,
} from '@angular/core';
import Globe from 'globe.gl';
import {fromEvent, Subscription} from 'rxjs';
import {Router} from '@angular/router';
import {ComponentPortal, DomPortalOutlet} from '@angular/cdk/portal';
import {StatusDotComponent} from '../status-dot';
import {DeploymentTarget} from '@glasskube/cloud-sdk';

type ValueOrPredicate<IN, OUT> = OUT | ((arg: IN) => OUT);

@Component({
  selector: 'app-globe',
  imports: [],
  template: ``,
})
export class GlobeComponent implements OnInit, OnChanges, OnDestroy {
  @Input({required: true}) public deploymentTargets!: DeploymentTarget[];
  @Input() public dotRadius: ValueOrPredicate<DeploymentTarget, number> = 1;
  @Input() public labelSize: ValueOrPredicate<DeploymentTarget, number> = 3.5;

  private readonly router = inject(Router);
  private readonly hostElement = inject(ElementRef).nativeElement as HTMLElement;
  private readonly parentElement = this.hostElement.parentElement;
  private readonly globeInstance = new Globe(this.hostElement);
  private readonly resize$ = fromEvent(window, 'resize');
  private readonly subscription = new Subscription();
  private readonly app = inject(ApplicationRef);

  private get parentHeight(): number {
    return this.parentElement?.offsetHeight ?? 0;
  }

  private get parentWidth(): number {
    return this.parentElement?.offsetWidth ?? 0;
  }

  public ngOnInit(): void {
    this.initGlobe();
    this.subscription.add(this.resize$.subscribe(() => this.updateDimensions()));
  }

  public ngOnChanges(changes: SimpleChanges): void {
    if (changes['deploymentTargets']) {
      this.updateLabelData();
    }
    if (changes['labelSize']) {
      this.updateLabelSize();
    }
    if (changes['dotRadius']) {
      this.updateDotRadius();
    }
  }

  public ngOnDestroy(): void {
    this.subscription.unsubscribe();
    this.globeInstance._destructor();
  }

  private initGlobe(): void {
    this.updateDimensions();
    this.updateLabelSize();
    this.updateDotRadius();
    // Globe reference:
    // https://github.com/vasturiano/three-globe/#api-reference
    this.globeInstance
      .globeImageUrl(
        // Other options: https://github.com/vasturiano/three-globe/tree/master/example/img
        'https://raw.githubusercontent.com/vasturiano/three-globe/refs/heads/master/example/img/earth-blue-marble.jpg'
      )
      .backgroundColor('rgba(0,0,0,0)')
      .atmosphereColor('rgba(0,0,0,0)')
      .showGraticules(false)
      .htmlElement((dt) => this.getMarkerElement(dt as DeploymentTarget))
      .htmlLat((dt) => (dt as DeploymentTarget).geolocation!.lat)
      .htmlLng((dt) => (dt as DeploymentTarget).geolocation!.lon);
  }

  private updateLabelData(): void {
    this.globeInstance.htmlElementsData(this.deploymentTargets.filter((dt) => dt.geolocation));
  }

  private updateLabelSize(): void {
    this.globeInstance.labelSize(this.labelSize as any);
  }

  private updateDotRadius(): void {
    this.globeInstance.labelDotRadius(this.dotRadius as any);
  }

  private updateDimensions(): void {
    const d = Math.min(this.parentHeight, this.parentWidth);
    this.globeInstance.width(d).height(d);
  }

  private getMarkerElement(dt: DeploymentTarget): HTMLElement {
    const el = document.createElement('a');
    el.classList.add('flex', 'flex-col', 'items-center', 'text-blue-100', 'cursor-pointer');
    // This makes the element clickable
    el.style.setProperty('pointer-events', 'auto');
    // Set a href, so the destination shows up in the browser chrome
    el.href = '/deployment-targets';
    // Navigate to the destination via angular routing to prevent page reload.
    el.addEventListener('click', (event) => {
      event.preventDefault();
      return this.navigateToDeploymentTargets();
    });
    el.innerHTML = `
      <div class="relative">
        <img src="/docker.png" alt="docker" class="w-6 h-6 rounded">
        <div class="dot-outlet absolute w-2 h-2 -bottom-0.5 -end-0.5"></div>
      </div>
      ${dt.name}`;
    const cp = new ComponentPortal(StatusDotComponent);
    const ref = cp.attach(new DomPortalOutlet(el.querySelector('.dot-outlet')!, undefined, this.app));
    ref.instance.deploymentTarget = dt;
    return el;
  }

  private async navigateToDeploymentTargets(): Promise<void> {
    await this.router.navigate(['/deployment-targets']);
  }
}
