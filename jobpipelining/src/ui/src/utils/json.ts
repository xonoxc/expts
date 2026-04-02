import { fromThrowable } from "neverthrow"

export function SafeParseJSON(data: string) {
   return fromThrowable(JSON.parse, () => ({ message: "Failed to parse JSON" }))(data)
}

export function SafeStringifyJSON<T>(data: T) {
   return fromThrowable(JSON.stringify, () => ({
      message: "Failed to parse JSON",
   }))(data)
}
