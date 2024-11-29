import {Pipe, PipeTransform} from '@angular/core';
import dayjs from 'dayjs';
import {Duration} from 'dayjs/plugin/duration';

export function isOlderThan(date: dayjs.ConfigType, duration: Duration): boolean {
  return dayjs.duration(Math.abs(dayjs(date).diff(dayjs()))) > duration;
}

@Pipe({standalone: true, name: 'relativeDate'})
export class RelativeDatePipe implements PipeTransform {
  transform(value: dayjs.ConfigType, withoutSuffix: boolean = false): string {
    return dayjs(value).toNow(withoutSuffix);
  }
}
