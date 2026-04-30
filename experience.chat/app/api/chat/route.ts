import { generateText } from "ai"
import { createGroq } from "@ai-sdk/groq"
import { attempt, attemptSync } from "@/lib/utils/attempt"
import { systemPrompt } from "@/consts/prompt"

import type { Chunk } from "@/lib/chunk-schema"

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

   const stream = new ReadableStream({
      async start(controller) {
         for (let i = 0; i < chunks.length; i++) {
            const chunk = chunks[i]
            controller.enqueue(`data: ${JSON.stringify(chunk)}\n\n`)

            if (i < chunks.length - 1) {
               await new Promise(resolve => setTimeout(resolve, 10))
            }
         }
         controller.enqueue(`data: [DONE]\n\n`)
         controller.close()
      },
   })

   return new Response(stream, {
      headers: {
         "Content-Type": "text/event-stream",
         "Cache-Control": "no-cache",
         Connection: "keep-alive",
      },
   })
}
