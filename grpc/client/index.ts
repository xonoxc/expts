import * as grpc from "@grpc/grpc-js"
import { MetadataServiceClient } from "./generated/metadata"

import type { MetadataRequest } from "./generated/metadata"

const client = new MetadataServiceClient(
  "localhost:50051",
  grpc.credentials.createInsecure(),
) 

const request: MetadataRequest = {key: "hello"}


client.getMetadata(request , (error, resp) => {
	console.log("getMetadata called with request:", request)

	 if(error){
		console.error(error)
		return
	}
	console.log(resp)
})
