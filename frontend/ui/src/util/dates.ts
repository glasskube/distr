import {Pipe, PipeTransform} from '@angular/core';
import dayjs from 'dayjs';
import {Duration} from 'dayjs/plugin/duration';

export function isOlderThan(date: dayjs.ConfigType, duration: Duration): boolean {
  return dayjs.duration(Math.abs(dayjs(date).diff(dayjs()))) > duration;
}

@Pipe({name: 'relativeDate'})
export class RelativeDatePipe implements PipeTransform {
  transform(value: dayjs.ConfigType, withoutSuffix: boolean = false): string {
    const d = dayjs(value);
    if (d.isBefore()) {
      return dayjs(value).fromNow(withoutSuffix);
    } else {
      return dayjs(value).toNow(withoutSuffix);
    }
  }
}
