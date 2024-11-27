import {BaseModel, Named} from '../app/types/base';

export function distinctById<T extends BaseModel>(input: readonly T[]): T[] {
  return input.filter((value: T, index, self) => {
    return self.findIndex((element) => element.id === value.id) === index;
  });
}

export function compareBy<T>(predicate: (arg: T) => string): (a: T, b: T) => number {
  return (a, b) => predicate(a).localeCompare(predicate(b));
}

export const compareByName = compareBy((arg: Named) => arg.name ?? '');
