export function never(arg: never): never {
  throw new Error(`Unexpected value: ${arg}`);
}
