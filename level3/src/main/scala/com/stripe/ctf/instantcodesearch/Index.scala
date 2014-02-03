package com.stripe.ctf.instantcodesearch

import java.io._
import scala.collection.mutable.HashMap
import scala.collection.mutable.ArrayBuffer

class Index(repoPath: String) {
  var idx = HashMap.empty[String, ArrayBuffer[Match]]
  var fileCounter = 0

  def path() = repoPath

  def addFile(file: String, text: String) {
    val beginningTime = System.currentTimeMillis
    fileCounter += 1

    text.split("[\r\n]+").zipWithIndex.foreach { case(line, lineNumber) =>
      var m = new Match(file, lineNumber + 1)
      line.split("\\s+").foreach { case(word) =>
        if(!word.isEmpty()) {
          if(idx.contains(word)) {
            var buffer = idx.apply(word)
            idx.update(word, buffer += m)
          } else {
            var b = new ArrayBuffer[Match]
            b += m
            idx.put(word, b)
          }
        }
      }
      fileCounter += 1
    }

    val time = (System.currentTimeMillis - beginningTime)
    if (time > 1000) {
      println("Slow index (" + time + "ms|" + fileCounter + "): \t" + file)
    }
  }
}

