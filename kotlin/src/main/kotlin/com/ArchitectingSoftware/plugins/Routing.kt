package com.ArchitectingSoftware.plugins

import com.ArchitectingSoftware.ChainDefaults
import com.ArchitectingSoftware.RequestBCParams
import com.ArchitectingSoftware.SolveHash
import com.ArchitectingSoftware.BlockSolution
import io.ktor.server.routing.*
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.response.*
import io.ktor.server.request.*

import io.ktor.serialization.kotlinx.json.*
import kotlinx.serialization.*
import kotlinx.serialization.json.*

fun Application.configureRouting() {

    // Starting point for a Ktor app:
    routing {
        get("/") {
            call.respondText("Hello World!")
        }

        get("/bc") {
            val qp = call.request.queryParameters

            println("Solving Block ID: " + qp["b"])
            val p = RequestBCParams(
                q = qp["q"] ?: ChainDefaults.reqQuery,
                b= qp["b"] ?: ChainDefaults.reqBlockID,
                m= qp["m"]?.toLong() ?: ChainDefaults.reqMaxTries,
                x= qp["x"] ?: ChainDefaults.signPrefix,
                p= qp["p"] ?: ChainDefaults.nullHash)

            val msg = SolveHash(p)
            call.respond(msg)
        }
    }
    routing {
    }
}
