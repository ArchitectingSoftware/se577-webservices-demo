package com.ArchitectingSoftware

import io.ktor.server.engine.*
import io.ktor.server.netty.*
import io.ktor.server.routing.*
import io.ktor.server.response.*
import com.ArchitectingSoftware.plugins.*
import io.ktor.server.application.*

fun main() {
    embeddedServer(Netty, port = 9096, host = "0.0.0.0") {
        configureHTTP()
        configureMonitoring()
        configureContentNegotiation()
        configureRouting()
    }.start(wait = true)
}
