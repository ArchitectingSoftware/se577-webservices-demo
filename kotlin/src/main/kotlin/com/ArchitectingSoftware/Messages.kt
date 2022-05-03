package com.ArchitectingSoftware

import io.ktor.serialization.kotlinx.json.*
import kotlinx.serialization.*
import kotlinx.serialization.json.*

object ChainDefaults {
    val zeroHash = "0000000000000000000000000000000000000000000000000000000000000000"
    val nullHash = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
    val signPrefix = "0000"
    val reqQuery = "HELLO WORLD!"
    val reqBlockID = "12345"
    val reqMaxTries = 1000000L
}

@Serializable
data class BlockSolution (
    val blockHash: String = ChainDefaults.nullHash,
    val blockId: String = "",
    val executionTimeMs: Long = 0L,
    val found : Boolean = false,
    val nonce: Long = 0L,
    val parentHash : String = ChainDefaults.zeroHash,
    val query : String = ""
)

data class RequestBCParams(
    val q: String,
    val p: String,
    val b: String,
    val m: Long,
    val x:String = ChainDefaults.signPrefix
)


/*
"blockHash": "000faa760498b8a830f5d4c0f7a456652c675687212fa8ca025e90be7d8bf84e",
"blockId": "e51196cf-2ce7-4eff-8a62-3f158d07788e",
"executionTimeMs": 4,
"found": true,
"nonce": 6259,
"parentHash": "00000000000000000000000000000000",
"query": "scala"
*/
