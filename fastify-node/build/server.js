"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fastify_1 = require("fastify");
const swagger_1 = __importDefault(require("@fastify/swagger"));
const fastify_cors_1 = __importDefault(require("fastify-cors"));
const fastify_http_proxy_1 = __importDefault(require("fastify-http-proxy"));
const schemas_1 = require("./schemas");
const dotenv_1 = __importDefault(require("dotenv"));
const crypto_js_1 = __importDefault(require("crypto-js"));
const operators_1 = require("rxjs/operators");
const rxjs_1 = require("rxjs");
// Read in secret key via the .env file.  This will not be checked into github becuase of
// .gitignore.   Create a file called .env and make sure it has the following line:
//
//  GITHUB_ACCESS_TOKEN = "YOUR_PERSONAL_ACCESS_KEY_FROM_GITHUB"
dotenv_1.default.config();
// create the fastify task
const server = (0, fastify_1.fastify)({
    logger: true
});
//register the swagger middleware, listen on /docs
server.register(swagger_1.default, {
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
});
//setup a proxy to github, inject the authoorization header
server.register(fastify_http_proxy_1.default, {
    upstream: 'https://api.github.com',
    prefix: 'ghsecure',
    httpMethods: ['GET', 'POST'],
    replyOptions: {
        rewriteRequestHeaders: (origReq, headers) => {
            return Object.assign(Object.assign({}, headers), { authorization: `Token ${process.env.GITHUB_ACCESS_TOKEN}` });
        }
    }
});
//just a dumb proxy for demo purposoes
server.register(fastify_http_proxy_1.default, {
    upstream: 'https://api.github.com',
    httpMethods: ['GET', 'POST'],
    prefix: 'gh',
});
//setup CORS 
server.register(fastify_cors_1.default, {
    origin: "*"
});
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
        querystring: schemas_1.BcQueryStringSchema,
        response: {
            200: schemas_1.BcResultSchema,
            400: { type: 'string' }
        }
    }
};
server.get('/bc', bcAPIConfig, async (request, reply) => {
    //destructure the command line args
    let { q, p, b, x, m } = request.query;
    let respObj = {
        blockHash: '',
        blockId: b,
        executionTimeMs: 0,
        found: false,
        nonce: 0,
        parentHash: p,
        query: q
    };
    //shouldnt happen because we have defaults
    if ((x == undefined) || (m == undefined)) {
        reply.code(400).send("Error required values are undefined");
        return;
    }
    const baseHashStr = b + q + p;
    const startTime = new Date().getTime();
    let found = false;
    for (let i = 0; i <= m; i++) {
        const hValue = crypto_js_1.default.SHA256(baseHashStr + i).toString();
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
    if (found === false) {
        respObj.blockHash = crypto_js_1.default.SHA256(baseHashStr + m).toString();
        respObj.nonce = m;
    }
    reply.code(200).send(respObj);
});
server.get('/bc2', bcAPIConfig, async (request, reply) => {
    //destructure the command line args
    let { q, p, b, x, m } = request.query;
    //build the base response object, key values will be updated below
    let respObj = {
        blockHash: '',
        blockId: b,
        executionTimeMs: 0,
        found: false,
        nonce: 0,
        parentHash: p,
        query: q
    };
    //shouldnt happen because we have defaults
    if ((x === undefined) || (m === undefined)) {
        reply.code(400).send("Error required values are undefined");
        return;
    }
    //the json schema utility types x and m as possibly being undefined because they
    //are optional, at this point they cannot be based on the check above, so setting
    //up new variables to drop the undefined type option
    const complexity = x;
    const maxTries = m;
    const baseHashStr = b + q + p;
    const startTime = new Date().getTime();
    let found = false;
    let result = (0, rxjs_1.range)(0, maxTries).pipe((0, operators_1.find)(i => {
        const hValue = crypto_js_1.default.SHA256(baseHashStr + i).toString();
        if (hValue.startsWith(complexity)) {
            return true;
        }
        else {
            return false;
        }
    })).pipe((0, operators_1.switchMap)(i => {
        // i will be undefined if the value was not found
        if (i) {
            respObj.found = true;
            respObj.nonce = i;
        }
        else {
            respObj.found = false;
            respObj.nonce = maxTries;
        }
        respObj.blockHash = crypto_js_1.default.SHA256(baseHashStr + respObj.nonce).toString();
        respObj.executionTimeMs = new Date().getTime() - startTime;
        //wrap the response object into an observable so that it can be picked up in the
        //subscribe - thats what of() does
        return (0, rxjs_1.of)(respObj);
    })).subscribe((response => reply.code(200).send(response)));
});
//set the server up, just hard coding the listen port for now to 4080
server.listen(9094, "::", (err, address) => {
    if (err) {
        console.error(err);
        process.exit(1);
    }
    console.log(`Server listening at ${address}`);
});
//# sourceMappingURL=server.js.map