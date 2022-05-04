import { fastify, RouteShorthandOptions } from 'fastify'
import fastifySwagger from '@fastify/swagger'
import fastifyCors from 'fastify-cors'
import fastifyHttpProxy from 'fastify-http-proxy'
import { FromSchema } from "json-schema-to-ts"
import { BcQueryStringSchema, BcResultSchema} from './schemas'
import dotenv from 'dotenv'
import {v4 as uuidv4} from 'uuid'
import CryptoJS from 'crypto-js'

import {map,first, switchMap, findIndex, find} from 'rxjs/operators';
import { Observable, Subject, Subscription, range, from, of, defaultIfEmpty,isEmpty} from 'rxjs';


// Read in secret key via the .env file.  This will not be checked into github becuase of
// .gitignore.   Create a file called .env and make sure it has the following line:
//
//  GITHUB_ACCESS_TOKEN = "YOUR_PERSONAL_ACCESS_KEY_FROM_GITHUB"
dotenv.config()

// create the fastify task
const server = fastify({
    logger: true
})

//register the swagger middleware, listen on /docs
server.register(fastifySwagger, {
    exposeRoute: true,
    routePrefix: '/docs',
    swagger: {
        info: {
            title: 'BC Testing Swagger',
            description: 'Testing the swagger API',
            version: '0.1.0'
        },
        host: 'localhost:4080',
        schemes: ['http'],
        consumes: ['application/json'],
        produces: ['application/json']
    },
    
})

//setup a proxy to github, inject the authoorization header
server.register (fastifyHttpProxy, {
    upstream: 'https://api.github.com',
    prefix: 'ghsecure', 
    httpMethods: ['GET', 'POST'],
    replyOptions: {
        rewriteRequestHeaders: (origReq, headers) => {
            return {
                ...headers,
                authorization: `Token ${process.env.GITHUB_ACCESS_TOKEN}`
            }
        } 
    } 
})

//just a dumb proxy for demo purposoes
server.register (fastifyHttpProxy, {
    upstream: 'https://api.github.com',
    httpMethods: ['GET', 'POST'],
    prefix: 'gh', 
})

//setup CORS 
server.register(fastifyCors, {
    origin: "*"
})

/*
interface BCResponse   {
    blockHash: string;
    blockId: string;
    executionTimeMs: number;
    found: boolean;
    nonce: number;
    parentHash: string;
    query: string;
} */

const bcAPIConfig = {
    schema: {
        querystring: BcQueryStringSchema,
        response : {
            200: BcResultSchema,
            400: {type: 'string'}
        }
    }
}

server.get<{
        Querystring: FromSchema<typeof BcQueryStringSchema>,
        Reply: FromSchema<typeof BcResultSchema> | string, // BCResponse
    }>('/bc', bcAPIConfig, async (request, reply) => {
    
    //destructure the command line args
    let {q,p,b,x,m} = request.query;
    let respObj : FromSchema<typeof BcResultSchema> ={
        blockHash: '',
        blockId: b,
        executionTimeMs: 0,
        found: false,
        nonce: 0,
        parentHash: p,
        query:  q
    }
    

    //shouldnt happen because we have defaults
    if ((x == undefined) || (m == undefined)){
        reply.code(400).send("Error required values are undefined")
        return;
    }

    const baseHashStr = b + q + p;
    const startTime = new Date().getTime();
    let found = false;

    for (let i = 0; i <= m; i++) {
        const hValue = CryptoJS.SHA256(baseHashStr + i).toString();
        if (hValue.startsWith(x)) {
            found = true;
            respObj.blockHash = hValue;
            respObj.nonce = i;
            break;
        }
    }

    const currTime = new Date().getTime();
    respObj.executionTimeMs = currTime - startTime;
    respObj.found = found;

    if(found === false) {
        respObj.blockHash = CryptoJS.SHA256(baseHashStr + m).toString();
        respObj.nonce = m;
    }

    reply.code(200).send(respObj)
})


server.get<{
    Querystring: FromSchema<typeof BcQueryStringSchema>,
    Reply: FromSchema<typeof BcResultSchema> | string, // BCResponse
}>('/bc2', bcAPIConfig, async (request, reply) => {

//destructure the command line args
let {q,p,b,x,m} = request.query;

//build the base response object, key values will be updated below
let respObj : FromSchema<typeof BcResultSchema> ={
    blockHash: '',
    blockId: b,
    executionTimeMs: 0,
    found: false,
    nonce: 0,
    parentHash: p,
    query:  q
}


//shouldnt happen because we have defaults
if ((x === undefined) || (m === undefined)){
    reply.code(400).send("Error required values are undefined")
    return;
}

//the json schema utility types x and m as possibly being undefined because they
//are optional, at this point they cannot be based on the check above, so setting
//up new variables to drop the undefined type option
const complexity: string = x
const maxTries: number = m

const baseHashStr = b + q + p;
const startTime = new Date().getTime();
let found = false;

let result = range(0, maxTries).pipe(
    find( i => {
        const hValue = CryptoJS.SHA256(baseHashStr + i).toString();
            if (hValue.startsWith(complexity)) {
              return true;
            } else {
              return false;
            }
    })).pipe(
        switchMap(i => {
            // i will be undefined if the value was not found
            if (i){
                respObj.found = true;
                respObj.nonce = i;
            } else {
                respObj.found = false;
                respObj.nonce = maxTries;
            }
            respObj.blockHash = CryptoJS.SHA256(baseHashStr + respObj.nonce).toString();
            respObj.executionTimeMs =  new Date().getTime() - startTime;
            //wrap the response object into an observable so that it can be picked up in the
            //subscribe - thats what of() does
            return of(respObj)
        })
    ).subscribe(
        (response =>  reply.code(200).send(response))
    )
})


//set the server up, just hard coding the listen port for now to 4080
server.listen(4080, (err, address) => {
  if (err) {
    console.error(err)
    process.exit(1)
  }
  console.log(`Server listening at ${address}`)
})