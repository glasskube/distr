import {GlobalPositionStrategy, Overlay, OverlayConfig} from '@angular/cdk/overlay';
import {ComponentPortal, ComponentType, TemplatePortal} from '@angular/cdk/portal';
import {inject, Injectable, InjectionToken, Injector, TemplateRef, ViewContainerRef} from '@angular/core';
import {firstValueFrom, map, merge, Observable, Subject, take, takeUntil} from 'rxjs';
import {ConfirmDialogComponent} from '../components/confirm-dialog/confirm-dialog.component';

export class DialogRef<T = void> {
  private readonly result$ = new Subject<T>();
  private readonly dismissed$ = new Subject<null>();

  public close(data: T) {
    this.result$.next(data);
  }

  public dismiss(): void {
    this.dismissed$.next(null);
  }

  public closed(): Observable<void> {
    return this.result().pipe(map(() => undefined));
  }

  public result(): Observable<T | null> {
    return merge(this.result$, this.dismissed$).pipe(take(1));
  }
}

export class CurrentOverlay extends DialogRef<unknown> {}

export class ExtendedOverlayConfig extends OverlayConfig {
  backdropStyleOnly?: boolean;
  data?: unknown;
}

export const OverlayData = new InjectionToken<unknown>('OVERLAY_DATA');

@Injectable()
export class OverlayService {
  private readonly overlay = inject(Overlay);
  private readonly viewContainerRef = inject(ViewContainerRef);

  public confirm(message: string) {
    return this.showModal<boolean>(ConfirmDialogComponent, {data: {message}}).result();
  }

  /**
   * @param templateRef the template to show
   * @param viewContainerRef needed to create a TemplatePortal. you can get it by injecting `ViewContainerRef`
   * @param config optional overlay config
   * @returns a handle of the modal with some control functions
   */
  public showModal<T = void>(
    templateRefOrComponentType: TemplateRef<unknown> | ComponentType<unknown>,
    config?: ExtendedOverlayConfig
  ): DialogRef<T> {
    return this.show(templateRefOrComponentType, {
      hasBackdrop: true,
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().top(),
      ...config,
    });
  }

  public showDrawer<T = void>(templateRefOrComponentType: TemplateRef<unknown> | ComponentType<unknown>): DialogRef<T> {
    return this.show(templateRefOrComponentType, {
      hasBackdrop: true,
      positionStrategy: new GlobalPositionStrategy().end().centerVertically(),
    });
  }

  private show<T>(
    templateRefOrComponentType: TemplateRef<unknown> | ComponentType<unknown>,
    config: ExtendedOverlayConfig
  ): DialogRef<T> {
    const overlayRef = this.overlay.create(config);
    const dialogRef = new DialogRef<T>();
    const injector =
      config.data !== null && config.data !== undefined
        ? Injector.create({
            parent: this.viewContainerRef.injector,
            providers: [
              {provide: DialogRef, useValue: dialogRef},
              {provide: OverlayData, useValue: config.data},
            ],
          })
        : null;

    if (templateRefOrComponentType instanceof TemplateRef) {
      const embeddedViewRef = overlayRef.attach(
        new TemplatePortal(templateRefOrComponentType, this.viewContainerRef, injector)
      );
      dialogRef.closed().subscribe(() => embeddedViewRef.destroy());
    } else {
      const componentRef = overlayRef.attach(new ComponentPortal(templateRefOrComponentType, null, injector));
      dialogRef.closed().subscribe(() => componentRef.destroy());
    }

    if (!config.backdropStyleOnly) {
      overlayRef
        .backdropClick()
        .pipe(takeUntil(dialogRef.closed()))
        .subscribe(() => dialogRef.dismiss());
    }

    return dialogRef;
  }
}
