import {Component, computed, input} from '@angular/core';

@Component({
  selector: 'app-quota-limit',
  template: `
    @let l = limit();
    @let p = percentage();
    @if (l !== undefined && p >= 75) {
      <div class="flex text-gray-900 dark:text-white flex-grow items-center justify-end gap-2">
        <div class="bg-gray-200 rounded-full h-2.5 dark:bg-gray-700 max-w-24 flex-grow">
          <div
            class="h-2.5 rounded-full"
            [class.bg-blue-600]="!isLimitCritical() && !isLimitReached()"
            [class.bg-yellow-400]="isLimitCritical() && !isLimitReached()"
            [class.bg-red-600]="isLimitReached()"
            [class.dark:bg-red-500]="isLimitReached()"
            [style]="{width: p + '%'}"></div>
        </div>
        <span class="text-sm" [class.text-red-700]="isLimitReached()" [class.dark:text-red-500]="isLimitReached()">
          {{ usage() ?? 0 }}/{{ l }}
        </span>
      </div>
    }
  `,
})
export class QuotaLimitComponent {
  public readonly usage = input<number>();
  public readonly limit = input<number>();
  protected readonly percentage = computed(() => {
    const u = this.usage();
    const l = this.limit();
    if (!l || l < 0) {
      return 0;
    }
    return Math.min(100, Math.round(((u ?? 0) / l) * 100));
  });
  public isLimitCritical = computed(() => this.percentage() >= 85);
  public isLimitReached = computed(() => this.percentage() >= 100);
}
