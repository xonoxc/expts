import type { Context, Next } from "hono"
import { attempt } from "./ui/src/utils/attempt"

export function betterLogger() {
	return async (c: Context, next: Next) => {
		const start = Date.now()
		const requestId = crypto.randomUUID()

		c.set("requestId", requestId)

		console.log(`[REQ] ${requestId} → ${c.req.method} ${c.req.path}`)

		const res = await attempt(next())
		if(res.isErr()){
			const duration = Date.now() - start
			console.error(`[ERR] ${requestId} ✖ ${c.req.method} ${c.req.path} (${duration}ms)`)
			console.error(res.error)

			throw res.error
		}

		const duration = Date.now() - start

		console.log(`[RES] ${requestId} ← ${c.res.status} (${duration}ms)`)
	}
}
