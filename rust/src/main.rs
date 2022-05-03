use actix_web::{get, post, web, App, HttpResponse, HttpServer, Responder};
use serde::{Serialize,Deserialize};
use hex_literal::hex;
use hex::encode;
use sha2::{Sha256, Digest};
use std::time::{Duration,Instant};
use actix_cors::Cors;
use numtoa::NumToA;
use faster_hex::{hex_string};

#[derive(Deserialize)]
struct Info {
    q: String,  //query data
    p: String,  //parent hash
    b: String,  //block id
    #[serde(default="difficulty")]
    x: String,  //solver difficulity
    #[serde(default="max_default")]
    m: u64      //max iterations
}

#[derive(Serialize)]
struct bc_result{
    blockHash: String,
    nonce: u64,
    found: bool,
    parentHash: String,
    blockId: String,
    query: String,
    executionTimeMs: u128
}

const nullHash: &'static str = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF";

//Setup default handlers
fn difficulty() -> String {
    return "000".to_string();
}

fn max_default() -> u64 {
    return 1000000;
}

#[get("/bc")]
async fn bc_handler(params: web::Query<Info>) -> impl Responder {
    println!("Parameters {}, {}, {}, {}, {}", params.q, 
        params.p, params.b, params.x, params.m);

    let start = Instant::now();
    let mut shax = Sha256::new();
    

    for i in 0..params.m{
        let mut sha = shax.to_owned();

        sha.reset();
        let input = format!("{}{}{}{}",&params.b,&params.q,&params.p,i);
        sha.update(input.as_bytes());
        let result = hex::encode(sha.finalize());

        if result.starts_with(&params.x) {
            let duration=start.elapsed();

            let res = bc_result{
                blockHash: result,
                nonce: i,
                found:true,
                parentHash: params.p.to_owned(),
                blockId: params.b.to_owned(),
                query: params.q.to_owned(),
                executionTimeMs: duration.as_millis()
            };
            let j = serde_json::to_string(&res).unwrap();

            return HttpResponse::Ok()
                .content_type("application/json")
                .body(j);
            //break;
        }
   }
   //We didnt find it
   let duration=start.elapsed();    
   let res = bc_result{
        blockHash: nullHash.to_string(),
        nonce: params.m,
        found:false,
        parentHash: params.p.to_owned(),
        blockId: params.b.to_owned(),
        query: params.q.to_owned(),
        executionTimeMs: duration.as_millis()
    };
    let j = serde_json::to_string(&res).unwrap();

    return HttpResponse::Ok()
        .content_type("application/json")
        .body(j);
}

#[get("/bc2")]
async fn bc_handler2(params: web::Query<Info>) -> impl Responder {
    println!("Parameters2 {}, {}, {}, {}, {}", params.q, 
        params.p, params.b, params.x, params.m);

    let prefix = format!("{}{}{}", &params.b, &params.q, &params.p);
    let complexity = &params.x;

    let start = Instant::now();

    let hit = (0..params.m).find( |&i| {
        let input = format!("{}{}",prefix,i);
        let result = hex::encode(Sha256::digest(input.as_bytes()));
        result.starts_with(complexity)
    });

    let duration=start.elapsed();

    let res = match hit{
        None => {
            bc_result{
                blockHash: nullHash.to_string(),
                nonce: params.m.to_owned(),
                found:false,
                parentHash: params.p.to_owned(),
                blockId: params.b.to_owned(),
                query: params.q.to_owned(),
                executionTimeMs: duration.as_millis()
            }
        }
        Some(idx) => {
            let final_hash_input = format!("{}{}", prefix, idx);
            let result_hash = hex::encode(Sha256::digest(final_hash_input.as_bytes()));
            bc_result{
                blockHash: result_hash,
                nonce: idx,
                found:true,
                parentHash: params.p.to_owned(),
                blockId: params.b.to_owned(),
                query: params.q.to_owned(),
                executionTimeMs: duration.as_millis()
            }
        }
    };

    let j = serde_json::to_string(&res).unwrap();

    return HttpResponse::Ok()
        .content_type("application/json")
        .body(j);
}

//fastest
#[get("/bc3")]
async fn bc_handler3(params: web::Query<Info>) -> impl Responder {
    println!("Parameters3 {}, {}, {}, {}, {}", params.q, 
        params.p, params.b, params.x, params.m);

    let mut hasher = Sha256::new();
    let prefix_temp = format!("{}{}{}",&params.b,&params.q,&params.p);
    let prefix_len = prefix_temp.len();
    let mut prefix = String::with_capacity(prefix_temp.len() + 25);
    prefix.push_str(prefix_temp.as_str());
    let complexity = &params.x;
    let mut buf = [0u8; 20];

    let start = Instant::now();

    for i in 0..params.m{
        let inner_hasher = &mut hasher;
        let inner_index = &mut prefix;
        
        inner_index.truncate(prefix_len);
        inner_index.push_str(i.numtoa_str(10, &mut buf));

        inner_hasher.update(inner_index.as_bytes());

       let result = hex_string(inner_hasher.finalize_reset().as_slice());

        if result.starts_with(&params.x) {
            let duration=start.elapsed();

            let res = bc_result{
                blockHash: result,
                nonce: i,
                found:true,
                parentHash: params.p.to_owned(),
                blockId: params.b.to_owned(),
                query: params.q.to_owned(),
                executionTimeMs: duration.as_millis()
            };
            let j = serde_json::to_string(&res).unwrap();

            return HttpResponse::Ok()
                .content_type("application/json")
                .body(j);
        }
   }
   //We didnt find it
   let duration=start.elapsed();    
   let res = bc_result{
        blockHash: nullHash.to_string(),
        nonce: params.m,
        found:false,
        parentHash: params.p.to_owned(),
        blockId: params.b.to_owned(),
        query: params.q.to_owned(),
        executionTimeMs: duration.as_millis()
    };
    let j = serde_json::to_string(&res).unwrap();

    HttpResponse::Ok()
        .content_type("application/json")
        .body(j)
}


#[get("/")]
async fn hello() -> impl Responder {
    let start = Instant::now();
    let  testString = format!("{}", "hello world");
    let duration=start.elapsed(); 

    let rsp = format!("{}: {}", duration.as_millis(), testString);
    
    return HttpResponse::Ok().body(rsp);
}

#[post("/echo")]
async fn echo(req_body: String) -> impl Responder {
    HttpResponse::Ok().body(req_body)
}

async fn manual_hello() -> impl Responder {
    HttpResponse::Ok().body("Hey there!")
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .wrap(Cors::default().supports_credentials())
            //.wrap(Cors::new().supports_credentials().finish())
            .service(hello)
            .service(echo)
            .service(bc_handler)
            .service(bc_handler2)
            .service(bc_handler3)
            .route("/hey", web::get().to(manual_hello))
    })
    .bind("0.0.0.0:9099")?
    .run()
    .await
}