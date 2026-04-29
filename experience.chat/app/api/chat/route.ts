import { generateText } from "ai"
import { createGroq } from "@ai-sdk/groq"
import { attempt, attemptSync } from "@/lib/utils/attempt"

import type { Chunk } from "@/lib/chunk-schema"

const systemPrompt = `You are a helpful assistant that responds in structured chunks.
Each chunk must have: id (number), text (string), pace ("slow"|"normal"|"fast"), pauseAfter (number in seconds), and optional emphasis (boolean).
Return a JSON array of chunks.`

const groq = createGroq({
   apiKey: process.env.AI_API_KEY!,
})

export async function POST(req: Request) {
   const { messages } = await req.json()

   const result = await attempt(
      generateText({
         model: groq("openai/gpt-oss-20b"),
         system: systemPrompt,
         messages,
      })
   )

   if (result.isErr()) {
      const errorResponse = result.error
      console.log("AI response:", errorResponse)

      if (errorResponse.message.includes("Quota exceeded")) {
         return Response.json({
            type: "text",
            content: "AI quota exceeded. Please try again later.",
         })
      }

      return Response.json({
         type: "text",
         content: result.error instanceof Error ? result.error.message : "Unknown error from AI",
      })
   }

   const chunkRes = attemptSync<Chunk[]>(() => JSON.parse(result.value.output))
   if (chunkRes.isErr()) {
      console.error("Failed to parse AI response as JSON:", chunkRes.error)
      return Response.json({
         type: "text",
         content: "Malformed response from AI",
      })
   }

   const chunks = chunkRes.value

   if (!Array.isArray(chunks) || chunks.length === 0) {
      return Response.json({
         type: "text",
         content: "Failed to create response experience, falling back to message based rendering",
      })
   }

   return Response.json({ type: "chunks", content: chunks })
}
