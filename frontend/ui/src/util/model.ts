import dayjs from 'dayjs';
import {isOlderThan} from './dates';
import {Pipe, PipeTransform} from '@angular/core';
import {Duration} from 'dayjs/plugin/duration';
import {BaseModel} from '@glasskube/distr-sdk';

export function isStale(model: BaseModel, duration: Duration = dayjs.duration({seconds: 60})): boolean {
  return isOlderThan(model.createdAt, duration);
}

@Pipe({name: 'isStale'})
export class IsStalePipe implements PipeTransform {
  transform(value: BaseModel): boolean {
    return isStale(value);
  }
}
