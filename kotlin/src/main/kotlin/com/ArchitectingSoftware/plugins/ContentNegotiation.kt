package com.ArchitectingSoftware.plugins

import io.ktor.server.application.*
import io.ktor.server.plugins.contentnegotiation.*
import io.ktor.serialization.kotlinx.json.*

fun Application.configureContentNegotiation() {
    install(ContentNegotiation) {
        json()
    }
}