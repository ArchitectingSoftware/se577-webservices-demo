import * as Koa from 'koa';
import * as KoaRouter from 'koa-router';
import * as Crypto from 'crypto-js';
import * as cors from '@koa/cors'

var app = new Koa();
var router = new KoaRouter();

const NO_SHA_CONST = '99999999999999999999999999999999';
const NULL_SHA = '00000000000000000000000000000000';
const NO_SHA =   -1;
const SOLUTION_PREFIX = '000';

//setup host and port optionally via the environment
const NODE_PORT = +process.env.PORT || 9097;
const NODE_HOST = process.env.HOST || '0.0.0.0';

router.get('/bc', async http => {
    
    // process query parameters
    const q = http.query["q"];
    const p = http.query["p"];
    const b = http.query["b"];
    const x = http.query["x"];
    const m = http.query["m"];

    console.log(q, " ", p, " ", b, " ", x, " ", m)

    //const maxInter : number = 1000000;
    const maxInter: number = (m === undefined) ? 1000000 : m
    const solPrefix: number = (x === undefined) ? "000" : x

    if((q === undefined) || (p === undefined) || (b === undefined)){
        http.response.status = 422
        http.response.body = {
            message: "Invaid Parameter - Missing q, p or b"
        }
    } else
    {
        const baseHashStr = b + q + p;
        const startTime = new Date().getTime();
        let found = false;

        let respObj = {
            "blockHash": "000faa760498b8a830f5d4c0f7a456652c675687212fa8ca025e90be7d8bf84e",
            "blockId": b,
            "executionTimeMs": 0,
            "found": true,
            "nonce": 0,
            "parentHash": p,
            "query": q
        }

        for (let i = 0; i <= m; i++) {
            const hValue = Crypto.SHA256(baseHashStr + i).toString();
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
            respObj.blockHash = Crypto.SHA256(baseHashStr + m).toString();
            respObj.nonce = m;
        }

        http.response.body = respObj;
        console.log(respObj);
    }
  });


app
  .use(cors())
  .use(router.routes())
  .use(router.allowedMethods());


//app.use(async ctx => {
//    ctx.body = 'Hello World';
//  });
  
  app.listen(NODE_PORT);
  console.log('SERVER STARTED ON PORT: ' + NODE_PORT + '...');