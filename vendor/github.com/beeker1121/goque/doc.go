// Package goque provides embedded, disk-based implementations of stack, queue, and priority queue data structures.
//
// Motivation for creating this project was the need for a persistent priority queue that remained performant while growing well beyond the available memory of a given machine. While there are many packages for Go offering queues, they all seem to be memory based and/or standalone solutions that are not embeddable within an application.
//
// Instead of using an in-memory heap structure to store data, everything is stored using the Go port of LevelDB (https://github.com/syndtr/goleveldb). This results in very little memory being used no matter the size of the database, while read and write performance remains near constant.
//
// See README.md or visit https://github.com/beeker1121/goque for more info.
package goque
