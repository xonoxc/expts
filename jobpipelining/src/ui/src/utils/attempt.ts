import { fromPromise } from "neverthrow"

export function attempt<T>(promise: Promise<T>) {
   return fromPromise(promise, err => err)
}
