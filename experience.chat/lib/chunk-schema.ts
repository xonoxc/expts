import { z } from "zod/v4"

export type Chunk = {
  id: number
  text: string
  pace: "slow" | "normal" | "fast"
  pauseAfter: number
  emphasis?: boolean
}

export const chunkSchema = z.object({
  id: z.number(),
  text: z.string(),
  pace: z.enum(["slow", "normal", "fast"]),
  pauseAfter: z.number(),
  emphasis: z.boolean().optional()
})
