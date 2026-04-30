"use client"

import { useState, useEffect, useRef } from "react"
import { attempt } from "@/lib/utils/attempt"
import { usePlaybackStore } from "@/store/playback-store"
import { AIMessage } from "@/components/AIMessage"
import { UserMessage } from "@/components/UserMessage"

import type { Chunk } from "@/lib/chunk-schema"

export default function Page() {
   const [input, setInput] = useState("")
   const [userMessages, setUserMessages] = useState<Array<{ content: string }>>([])
   const [loading, setLoading] = useState(false)
   const [error, setError] = useState<string | null>(null)
   const messagesEndRef = useRef<HTMLDivElement>(null)
   const { reset, chunks, currentIndex } = usePlaybackStore()

   async function sendMessage() {
      if (!input.trim()) return

      const userMessage = input.trim()
      setUserMessages(prev => [...prev, { content: userMessage }])
      setInput("")
      reset()
      setLoading(true)
      setError(null)

      const result = await attempt(
         fetch("/api/chat", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
               messages: [
                  {
                     role: "user",
                     content: userMessage,
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

      const response = result.value

      if (!response.ok) {
         setError("Failed to reach API. Please try again.")
         setLoading(false)
         return
      }

      const reader = response.body?.getReader()
      const decoder = new TextDecoder()
      const collectedChunks: Chunk[] = []

      if (!reader) {
         setError("Failed to read response stream.")
         setLoading(false)
         return
      }

      let buffer = ""
      try {
         while (true) {
            const { done, value } = await reader.read()
            if (done) break

            buffer += decoder.decode(value, { stream: true })
            const lines = buffer.split("\n\n")
            buffer = lines.pop() || ""

            for (const line of lines) {
               if (line.startsWith("data: ")) {
                  const data = line.slice(6)
                  if (data === "[DONE]") {
                     break
                  }
                  try {
                     const chunk = JSON.parse(data)
                     collectedChunks.push(chunk)
                  } catch {
                     console.error("Failed to parse chunk:", data)
                  }
               }
            }
         }
      } catch (err) {
         console.error("Stream error:", err)
      }

      if (collectedChunks.length > 0) {
         usePlaybackStore.getState().setChunks(collectedChunks)
      } else {
         setError("No chunks received from AI")
      }

      setLoading(false)
   }

   useEffect(() => {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
   }, [userMessages, chunks, currentIndex])

   return (
      <div className="flex flex-col min-h-screen max-w-2xl mx-auto">
         <div className="flex-1 overflow-y-auto p-4 space-y-4">
            {userMessages.map((msg, i) => (
               <>
                  <UserMessage content={msg.content} />
                  {i === userMessages.length - 1 && <AIMessage />}
               </>
            ))}

            {loading && (
               <div className="flex justify-start mb-4">
                  <div className="bg-muted rounded-lg p-3 max-w-[80%] animate-pulse">
                     Thinking...
                  </div>
               </div>
            )}

            {error && (
               <div className="text-red-500 text-sm">
                  {error}
                  <button onClick={sendMessage} className="ml-2 underline">
                     Retry
                  </button>
               </div>
            )}

            <div ref={messagesEndRef} />
         </div>

         <div className="sticky bottom-0 bg-background p-4 border-t">
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
         </div>
      </div>
   )
}
