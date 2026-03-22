import type { Context, Next } from "hono"

export function betterLogger() {
   return async (c: Context, next: Next) => {
      const start = Date.now()
      const requestId = crypto.randomUUID()

      c.set("requestId", requestId)

      console.log(`[REQ] ${requestId} → ${c.req.method} ${c.req.path}`)

      try {
         await next()

         const duration = Date.now() - start

         console.log(`[RES] ${requestId} ← ${c.res.status} (${duration}ms)`)
      } catch (err) {
         const duration = Date.now() - start

         console.error(`[ERR] ${requestId} ✖ ${c.req.method} ${c.req.path} (${duration}ms)`)
         console.error(err)

         throw err
      }
   }
}
