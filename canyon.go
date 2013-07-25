// Copyright (c) 2013 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

/*
canyon is a tool for making big splits. This assumes that you have a very large
changelist prepared in a single branch. canyon will then split up the large
change into multiple branches, by OWNERS file, and then will prepare a changelist
description.
*/

var (
	maxDepth = flag.Int("depth", 0, "The maximum subdirectory depth for which split branches should be created.")

	upstreamBranch = flag.String("upstream", "origin/master", "The upstream branch against which diffs are taken and new branches created.")

	splitByType = flag.String("split-by", "[dir|file]", "The method by which the branch is split.")

	splitByFile = flag.String("split-by-file", "", "If using -split-by=file, this the common file name by which split directories are found.")
)

func main() {
	flag.Parse()

	if err := validateDescription(); err != nil {
		fmt.Println("Please provide a valid -message for your branches. Error:", err)
		flag.Usage()
		os.Exit(1)
	}

	if *splitByType != "dir" || *splitByType != "file" {
		fmt.Println("Invalid -split-by type:", *splitByType)
		flag.Usage()
		os.Exit(1)
	}
	if *splitByType == "file" && *splitByFile == "" {
		fmt.Println("Whe using -split-by=file, a -split-by-file is needed.")
		flag.Usage()
		os.Exit(1)
	}

	branch := strings.TrimSpace(gitOrDie("symbolic-ref", "--short", "HEAD"))

	fmt.Printf("Split changelist on branch %q into sub-changelists? [y/N] ", branch)
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	if buf[0] != 'y' {
		fmt.Println("Exiting")
		return
	}

	log.Print("Gathering changed files")
	files := strings.Split(gitOrDie("diff", "--name-only", *upstreamBranch), "\n")

	log.Print("Splitting changed files into groups for changelists")
	cs := newChangeSet(branch)
	for _, file := range files {
		if file == "" {
			continue
		}
		if *splitByType == "dir" {
			cs.splitByDir(file)
		} else if *splitByType == "file" {
			cs.splitByFile(*splitByFile, file)
		}
	}

	log.Print("Creating branches for splits")
	for _, cl := range cs.splits {
		splitBranch := cl.branchName(branch)
		log.Printf("Preparing branch %s", splitBranch)

		_, err := git("checkout", "-b", splitBranch, *upstreamBranch)
		if err != nil {
			log.Printf("Failed to create new branch %q: %v", splitBranch, err)
			continue
		}

		_, err = git("checkout", branch, cl.base)
		if err != nil {
			log.Print("Failed to check out subdirectory from root branch")
			gitOrDie("reset", "--hard", *upstreamBranch)
			continue
		}

		_, err = git("commit", "-a", "-m", formatDescription(cl))
		if err != nil {
			log.Print("Failed to create subchangelist")
			gitOrDie("reset", "--hard", *upstreamBranch)
			continue
		}
	}

	git("checkout", branch)
}

// git runs the specified git commands and returns the output as a string,
// blocking to completion.
func git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(stdout), nil
}

// gitOrDie runs the git command and panics on failure.
func gitOrDie(args ...string) string {
	r, e := git(args...)
	if e != nil {
		panic(e.Error())
	}
	return r
}
