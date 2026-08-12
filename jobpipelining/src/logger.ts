import type { Context, Next } from "hono"
import { attempt } from "./ui/src/utils/attempt"

export function betterLogger() {
	return async (ctx: Context, next: Next) => {
		const start = Date.now()
		const requestId = crypto.randomUUID()

		ctx.set("requestId", requestId)

		console.log(`[REQ] ${requestId} → ${ctx.req.method} ${ctx.req.path}`)

		const res = await attempt(next())
		if(res.isErr()){
			const duration = Date.now() - start
			console.error(`[ERR] ${requestId} ✖ ${ctx.req.method} ${ctx.req.path} (${duration}ms)`)
			console.error(res.error)

			throw res.error
		}

		const duration = Date.now() - start

		console.log(`[RES] ${requestId} ← ${ctx.res.status} (${duration}ms)`)
	}
}
