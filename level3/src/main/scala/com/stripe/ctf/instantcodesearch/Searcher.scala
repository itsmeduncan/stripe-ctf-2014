package com.stripe.ctf.instantcodesearch

import java.io._
import java.nio.file._

import com.twitter.concurrent.Broker

abstract class SearchResult
case class Match(path: String, line: Int) extends SearchResult
case class Done() extends SearchResult

class Searcher(index: Index)  {

  def search(needle : String, b : Broker[SearchResult]) = {
    for (m <- tryPath(index, needle)) {
      b !! m
    }

    b !! new Done()
  }

  def tryPath(index: Index, needle: String) : Iterable[SearchResult] = {
    try {
      return index.idx.filterKeys { _.contains(needle) }.values.flatten.toSet
    } catch {
      case e: IOException => {
        return Nil
      }
    }

    return Nil
  }

}
