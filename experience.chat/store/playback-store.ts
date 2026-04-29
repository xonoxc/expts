import { create } from "zustand"

type Chunk = {
   id: number
   text: string
   pace: "slow" | "normal" | "fast"
   pauseAfter: number
   emphasis?: boolean
}

type TimeoutId = ReturnType<typeof setTimeout> | null

type PlaybackState = {
   chunks: Chunk[]
   currentIndex: number
   isPlaying: boolean
   timeoutId: TimeoutId
   setChunks: (chunks: Chunk[]) => void
   play: () => void
   pause: () => void
   next: () => void
   prev: () => void
   reset: () => void
}

const paceMultipliers = {
   slow: 1.5,
   normal: 1,
   fast: 0.6,
} as const

function clearTimer(timeoutId: TimeoutId): null {
   if (timeoutId) clearTimeout(timeoutId)
   return null
}

export const usePlaybackStore = create<PlaybackState>((set, get) => ({
   chunks: [],
   currentIndex: 0,
   isPlaying: false,
   timeoutId: null,

   setChunks: chunks => set({ chunks, currentIndex: 0, isPlaying: true, timeoutId: null }),

   play: () => {
      const { chunks, currentIndex, isPlaying, timeoutId } = get()
      if (isPlaying || currentIndex >= chunks.length) return

      const chunk = chunks[currentIndex]
      const delay = chunk.pauseAfter * (paceMultipliers[chunk.pace] || 1) * 1000

      const newTimeoutId = setTimeout(() => {
         get().next()
      }, delay)

      set({ isPlaying: true, timeoutId: newTimeoutId })
   },

   pause: () =>
      set(state => ({
         isPlaying: false,
         timeoutId: clearTimer(state.timeoutId),
      })),

   next: () =>
      set(state => {
         const nextIndex = state.currentIndex + 1
         if (nextIndex >= state.chunks.length) {
            return {
               currentIndex: nextIndex,
               isPlaying: false,
               timeoutId: clearTimer(state.timeoutId),
            }
         }
         return {
            currentIndex: nextIndex,
            timeoutId: clearTimer(state.timeoutId),
         }
      }),

   prev: () =>
      set(state => ({
         currentIndex: Math.max(0, state.currentIndex - 1),
         timeoutId: clearTimer(state.timeoutId),
      })),

   reset: () =>
      set(state => ({
         currentIndex: 0,
         isPlaying: false,
         timeoutId: clearTimer(state.timeoutId),
      })),
}))
