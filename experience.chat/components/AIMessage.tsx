"use client"

import { useEffect } from "react"
import { motion } from "motion/react"
import { usePlaybackStore } from "@/store/playback-store"

export function AIMessage() {
  const { chunks, currentIndex, isPlaying, play, pause, next, reset } =
    usePlaybackStore()

  const isPlaybackComplete = currentIndex >= chunks.length

  useEffect(() => {
    if (chunks.length > 0 && currentIndex < chunks.length) {
      play()
    }
    return () => {
      pause()
    }
  }, [chunks, currentIndex])

  if (chunks.length === 0) return null

  return (
    <div className="flex justify-start mb-4">
      <div className="max-w-[80%] space-y-1">
        {chunks.map((chunk, index) => {
          const isCurrent = index === currentIndex

          return (
            <motion.div
              key={chunk.id}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.2 }}
              className={`${
                isPlaybackComplete
                  ? ""
                  : isCurrent
                  ? chunk.emphasis
                    ? "font-semibold"
                    : ""
                  : "opacity-40 blur-[1px]"
              }`}
            >
              {chunk.text}
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}
