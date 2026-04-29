import { attempt } from "@/lib/utils/attempt"
import { chunkSchema, type Chunk } from "@/lib/chunk-schema"

export type ParsedResult =
  | { type: "chunks"; content: Chunk[] }
  | { type: "text"; content: string }

const fallbackMessage = "Failed to create response experience, falling back to message based rendering"

export async function parseChunks(raw: string): Promise<ParsedResult> {
  const parseResult = await attempt(Promise.resolve().then(() => JSON.parse(raw)))

  if (parseResult.isErr()) {
    return { type: "text", content: fallbackMessage }
  }

  const data = parseResult.value
  const chunks = Array.isArray(data) ? data : [data]
  const validated: Chunk[] = []

  for (const item of chunks) {
    const result = chunkSchema.safeParse(item)
    if (result.success) {
      validated.push(result.data)
    }
  }

  if (validated.length === 0) {
    return { type: "text", content: fallbackMessage }
  }

  return { type: "chunks", content: validated }
}
