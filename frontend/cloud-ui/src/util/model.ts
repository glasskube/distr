import dayjs from 'dayjs';
import {BaseModel} from '../app/types/base';
import {isOlderThan} from './dates';
import {Pipe, PipeTransform} from '@angular/core';
import {Duration} from 'dayjs/plugin/duration';

export function isStale(model: BaseModel, duration: Duration = dayjs.duration({minutes: 5})): boolean {
  return isOlderThan(model.createdAt, duration);
}

@Pipe({name: 'isStale'})
export class IsStalePipe implements PipeTransform {
  transform(value: BaseModel): boolean {
    return isStale(value);
  }
}
