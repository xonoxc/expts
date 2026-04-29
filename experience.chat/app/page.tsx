"use client"

import { useState, useEffect } from "react"
import { attempt } from "@/lib/utils/attempt"
import { usePlaybackStore } from "@/store/playback-store"
import { ChunkPlayer } from "@/components/chunk-player"

import type { ParsedResult } from "@/lib/chunk-parser"

export default function Page() {
   const [input, setInput] = useState("")
   const [parsed, setParsed] = useState<ParsedResult | null>(null)
   const [loading, setLoading] = useState(false)
   const [error, setError] = useState<string | null>(null)
   const { reset } = usePlaybackStore()

   async function sendMessage() {
      if (!input.trim()) return

      reset()
      setLoading(true)
      setError(null)
      setParsed(null)

      const result = await attempt(
         fetch("/api/chat", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
               messages: [
                  {
                     role: "user",
                     content: input,
                  },
               ],
            }),
         })
      )

      if (result.isErr()) {
         setError("Failed to reach API. Please try again.")
         setLoading(false)
         return
      }

      const parsedResult: ParsedResult = await result.value.json()

      setParsed(parsedResult)
      setLoading(false)
   }

   useEffect(() => {
      if (parsed?.type === "chunks") {
         usePlaybackStore.getState().setChunks(parsed.content)
      }
   }, [parsed])

   return (
      <div className="max-w-2xl mx-auto p-4 space-y-4">
         <div className="flex gap-2">
            <input
               value={input}
               onChange={e => setInput(e.target.value)}
               onKeyDown={e => e.key === "Enter" && sendMessage()}
               className="flex-1 border rounded px-3 py-2"
               placeholder="Type a message..."
            />
            <button
               onClick={sendMessage}
               disabled={loading}
               className="px-4 py-2 bg-primary text-primary-foreground rounded disabled:opacity-50"
            >
               {loading ? <span className="animate-spin inline-block">⟳</span> : "Send"}
            </button>
         </div>

         {error && (
            <div className="text-red-500 text-sm">
               {error}
               <button onClick={sendMessage} className="ml-2 underline">
                  Retry
               </button>
            </div>
         )}

         {parsed?.type === "text" && (
            <p className="text-xs text-muted-foreground italic mt-1">{parsed.content}</p>
         )}

         {parsed?.type === "chunks" && <ChunkPlayer />}
      </div>
   )
}
