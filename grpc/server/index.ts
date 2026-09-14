import grpc from "@grpc/grpc-js"
import protoLoader from "@grpc/proto-loader" 

import { MetadataServiceService } from "./generated/metadata"

import type{MetadataServiceServer} from "./generated/metadata"


const packageDefinition  = protoLoader.loadSync("../proto/metadata.proto")

const grpcObject = grpc.loadPackageDefinition(packageDefinition)


const impl : MetadataServiceServer = {
	 getMetadata(call , callback) {
		console.log(call.request.key)

		callback(null , {
			 title : "hello from grpc",
			 statusCode : 200
		})
	}
}


const server = new grpc.Server()


server.addService(MetadataServiceService , impl)

server.bindAsync(
	'localhost:50051',
	grpc.ServerCredentials.createInsecure(),
	()=> {
		console.log("gRPC server running on :50051")
	}
)




console.log(grpcObject)

