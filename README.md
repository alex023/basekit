# basekit
The commonly used go language development kit

[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/alex023/basekit)](https://goreportcard.com/report/github.com/alex023/basekit)
[![GoDoc](https://godoc.org/github.com/alex023/basekit?status.svg)](https://godoc.org/github.com/alex023/basekit)
[![Build Status](https://travis-ci.org/alex023/basekit.svg)](https://travis-ci.org/alex023/basekit)

Some usual toolkit，include：
## pub-sub
Just mediate implemention in memory(publish：Topic，subscribe：channel）
## singleflight
only duplicate of singleflight in `groupcache`
## svc
program init,start,safe exit
## hash
- consistent:
- hrw:provides an implementation of Highest Random Weight hashing, an alternative to consistent hashing which is both simple and fast 
## other
- counter:multithreading counter
- waitwraper:a wrapper for simplify  calling for  waitgroup 