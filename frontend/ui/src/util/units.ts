import {formatNumber} from '@angular/common';
import {inject, LOCALE_ID, Pipe, PipeTransform} from '@angular/core';

const prefixes = ['', 'Ki', 'Mi', 'Gi'];

export function formatBytes(input: number, locale: string, digitsInfo?: string) {
  const index = Math.min(prefixes.length - 1, Math.floor(Math.log2(Math.abs(input)) / 10));
  return formatNumber(input / Math.pow(1024, index), locale, digitsInfo) + prefixes[index] + 'B';
}

@Pipe({name: 'bytes'})
export class BytesPipe implements PipeTransform {
  private readonly locale = inject(LOCALE_ID);

  transform(value: number, digitsInfo?: string) {
    return formatBytes(value, this.locale, digitsInfo);
  }
}
