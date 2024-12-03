import {GlobalPositionStrategy, Overlay, OverlayConfig} from '@angular/cdk/overlay';
import {TemplatePortal} from '@angular/cdk/portal';
import {EmbeddedViewRef, inject, Injectable, TemplateRef, ViewContainerRef} from '@angular/core';
import {Observable, Subject, takeUntil} from 'rxjs';

export class EmbeddedOverlayRef {
  private readonly closed$ = new Subject<void>();

  constructor(private readonly embeddedViewRef: EmbeddedViewRef<unknown>) {}

  public close() {
    this.embeddedViewRef.destroy();
    this.closed$.next();
  }

  public closed(): Observable<void> {
    return this.closed$;
  }
}

export class ExtendedOverlayConfig extends OverlayConfig {
  backdropStyleOnly?: boolean;
}

@Injectable({providedIn: 'root'})
export class OverlayService {
  private readonly overlay = inject(Overlay);

  /**
   * @param templateRef the template to show
   * @param viewContainerRef needed to create a TemplatePortal. you can get it by injecting `ViewContainerRef`
   * @param config optional overlay config
   * @returns a handle of the modal with some control functions
   */
  public showModal(
    templateRef: TemplateRef<unknown>,
    viewContainerRef: ViewContainerRef,
    config?: ExtendedOverlayConfig
  ): EmbeddedOverlayRef {
    return this.show(templateRef, viewContainerRef, {
      hasBackdrop: true,
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().top(),
      ...config,
    });
  }

  public showDrawer(templateRef: TemplateRef<unknown>, viewContainerRef: ViewContainerRef): EmbeddedOverlayRef {
    return this.show(templateRef, viewContainerRef, {
      hasBackdrop: true,
      positionStrategy: new GlobalPositionStrategy().end().centerVertically(),
    });
  }

  private show(
    templateRef: TemplateRef<unknown>,
    viewContainerRef: ViewContainerRef,
    config: ExtendedOverlayConfig
  ): EmbeddedOverlayRef {
    const overlayRef = this.overlay.create(config);
    const modalRef = new EmbeddedOverlayRef(overlayRef.attach(new TemplatePortal(templateRef, viewContainerRef)));
    if (!config.backdropStyleOnly) {
      overlayRef
        .backdropClick()
        .pipe(takeUntil(modalRef.closed()))
        .subscribe(() => modalRef.close());
    }
    return modalRef;
  }
}
