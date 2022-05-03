package com.ArchitectingSoftware

import java.security.MessageDigest
import kotlin.experimental.and

data class SolverTriple(val hash: String, val nonce:Long, val found: Boolean)

fun ByteArray.toHexString(): String{
    val hexArray = "0123456789ABCDEF".toCharArray()
    val result = StringBuffer()
    val hexChars = CharArray(this.size * 2)

    for(j in this.indices ){
        val v = (this[j] and 0xFF.toByte()).toInt()

        hexChars[j * 2] = hexArray[(v and 0xF0) ushr 4]
        hexChars[j * 2 + 1] = hexArray[v and 0x0F]
    }

    return String(hexChars)
}


fun SolveHash(req: RequestBCParams) : BlockSolution{

    val hasher = MessageDigest.getInstance("SHA-256")
    val blockBuffer = "${req.b}${req.q}${req.p}"
    val baseLen = blockBuffer.length
    val block = StringBuilder(baseLen+ 16).append(blockBuffer)


    val startTime = System.currentTimeMillis()
    val nonce = (0L..req.m).find { i ->
        val hval = hasher.digest(block.delete(baseLen, block.length).append(i).toString().toByteArray())
            .toHexString()
            .startsWith(req.x)
        hval
    }

    val result = when(nonce){
        is Long ->  {
            val finalBuffer = "${req.b}${req.q}${req.p}${nonce}"
            SolverTriple(hasher.digest(finalBuffer.toByteArray()).toHexString(), nonce,true)
        }
        else -> {
            val finalBuffer = "${req.b}${req.q}${req.p}${0L}"
            SolverTriple(hasher.digest(finalBuffer.toByteArray()).toHexString(), 0L, false)
        }
    }

    return BlockSolution(blockHash = result.hash,
        blockId = req.b,
        executionTimeMs = System.currentTimeMillis() - startTime,
        found = result.found,
        nonce = result.nonce,
        parentHash = req.p,
        query = req.q
    )
}